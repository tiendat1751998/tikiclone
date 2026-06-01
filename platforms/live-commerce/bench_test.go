package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/live-commerce/internal/fanout"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/websocket"
)

func BenchmarkJSONMarshalStdlib(b *testing.B) {
	msg := map[string]interface{}{
		"type":      "chat",
		"payload":   map[string]string{"user_id": "u1", "content": "hello world this is a test message"},
		"timestamp": time.Now().UnixMilli(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(msg)
	}
}

func BenchmarkJSONUnmarshalStdlib(b *testing.B) {
	data := []byte(`{"type":"chat","payload":{"user_id":"u1","content":"hello world"},"timestamp":1234567890}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg map[string]interface{}
		_ = json.Unmarshal(data, &msg)
	}
}

func BenchmarkSprintfKey(b *testing.B) {
	roomID := "room-12345"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("room:%s:viewers", roomID)
	}
}

func BenchmarkStringConcatKey(b *testing.B) {
	roomID := "room-12345"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = "room:" + roomID + ":viewers"
	}
}

func BenchmarkBroadcastFanout(b *testing.B) {
	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Register mock clients
	numClients := 100
	for i := 0; i < numClients; i++ {
		client := &websocket.Client{
			ID:     fmt.Sprintf("client-%d", i),
			RoomID: "test-room",
			send:   make(chan []byte, 256),
		}
		hub.Register <- client
	}

	broadcaster := fanout.NewBroadcaster(hub)
	defer broadcaster.Stop()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broadcaster.Broadcast(ctx, "test-room", "chat", map[string]interface{}{
			"user_id": "user1",
			"content": "test message for benchmarking",
		}, "")
	}
}

func BenchmarkBroadcastParallel(b *testing.B) {
	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Stop()

	numClients := 100
	for i := 0; i < numClients; i++ {
		client := &websocket.Client{
			ID:     fmt.Sprintf("client-%d", i),
			RoomID: "test-room",
			send:   make(chan []byte, 256),
		}
		hub.Register <- client
	}

	broadcaster := fanout.NewBroadcaster(hub)
	defer broadcaster.Stop()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			broadcaster.Broadcast(ctx, "test-room", "chat", map[string]interface{}{
				"user_id": "user1",
				"content": "test message",
			}, "")
			i++
		}
	})
}

func BenchmarkWebSocketHubRegister(b *testing.B) {
	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := &websocket.Client{
			ID:     fmt.Sprintf("client-%d", i),
			RoomID: "bench-room",
			send:   make(chan []byte, 256),
		}
		hub.Register <- client
	}
}

func BenchmarkBroadcastToRoom(b *testing.B) {
	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Stop()

	numClients := 50
	for i := 0; i < numClients; i++ {
		client := &websocket.Client{
			ID:     fmt.Sprintf("c%d", i),
			RoomID: "room-bench",
			send:   make(chan []byte, 256),
		}
		hub.Register <- client
	}

	data := []byte(`{"type":"chat","payload":{"content":"hello"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.BroadcastToRoomBytes("room-bench", data, "")
	}
}

func BenchmarkBroadcastToRoomParallel(b *testing.B) {
	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Stop()

	numClients := 50
	for i := 0; i < numClients; i++ {
		client := &websocket.Client{
			ID:     fmt.Sprintf("c%d", i),
			RoomID: "room-bench",
			send:   make(chan []byte, 256),
		}
		hub.Register <- client
	}

	data := []byte(`{"type":"chat","payload":{"content":"hello"}}`)
	var mu sync.Mutex
	_ = mu

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			hub.BroadcastToRoomBytes("room-bench", data, "")
		}
	})
}
