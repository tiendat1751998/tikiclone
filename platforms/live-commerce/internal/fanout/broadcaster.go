package fanout

import (
	"context"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/websocket"
	"go.uber.org/zap"
)

type BroadcastMsg struct {
	RoomID    string
	Event     string
	Payload   interface{}
	ExcludeID string
}

type Broadcaster struct {
	hub       *websocket.Hub
	input     chan *BroadcastMsg
	mu        sync.RWMutex
	global    map[string]chan struct{}
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	workerCnt int
}

func NewBroadcaster(hub *websocket.Hub) *Broadcaster {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Broadcaster{
		hub:       hub,
		input:     make(chan *BroadcastMsg, 4096),
		global:    make(map[string]chan struct{}),
		ctx:       ctx,
		cancel:    cancel,
		workerCnt: 4,
	}
	// Bounded worker pool: 4 workers process fanout to avoid unbounded goroutine growth
	for i := 0; i < b.workerCnt; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}
	return b
}

func (b *Broadcaster) worker(id int) {
	defer b.wg.Done()
	for {
		select {
		case msg, ok := <-b.input:
			if !ok {
				return
			}
			b.process(msg)
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *Broadcaster) process(msg *BroadcastMsg) {
	wrapper := map[string]interface{}{
		"type":      msg.Event,
		"payload":   msg.Payload,
		"timestamp": time.Now().UnixMilli(),
	}
	data, err := sonic.Marshal(wrapper)
	if err != nil {
		observability.LogWithTrace(nil).Error("broadcast marshal", zap.Error(err))
		return
	}
	b.hub.BroadcastToRoomBytes(msg.RoomID, data, msg.ExcludeID)
}

func (b *Broadcaster) Broadcast(ctx context.Context, roomID, event string, payload interface{}, excludeID string) {
	select {
	case b.input <- &BroadcastMsg{RoomID: roomID, Event: event, Payload: payload, ExcludeID: excludeID}:
	case <-ctx.Done():
		observability.LogWithTrace(ctx).Warn("broadcast cancelled",
			zap.String("room", roomID), zap.String("event", event))
	default:
		observability.LogWithTrace(ctx).Warn("broadcast buffer full",
			zap.String("room", roomID), zap.String("event", event))
	}
}

func (b *Broadcaster) BroadcastSync(roomID, event string, payload interface{}, excludeID string) {
	wrapper := map[string]interface{}{
		"type":      event,
		"payload":   payload,
		"timestamp": time.Now().UnixMilli(),
	}
	data, err := sonic.Marshal(wrapper)
	if err != nil {
		return
	}
	b.hub.BroadcastToRoomBytes(roomID, data, excludeID)
}

func (b *Broadcaster) RoomCount(roomID string) int {
	return b.hub.RoomCount(roomID)
}

func (b *Broadcaster) Stop() {
	b.cancel()
	close(b.input)
	b.wg.Wait()
}
