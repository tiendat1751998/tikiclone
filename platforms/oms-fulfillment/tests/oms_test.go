package oms_test

import (
	"context"
	"testing"
	"time"

	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/inventory"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/ordermanagement"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/pickpack"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/returns"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/warehouse"
)

func TestOrderLifecycle(t *testing.T) {
	repo := ordermanagement.NewInMemoryRepository()
	svc := ordermanagement.NewService(repo)

	order := &ordermanagement.Order{
		ID:     "order-001",
		UserID: "user-001",
		Items: []ordermanagement.OrderItem{
			{ID: "item-001", ProductID: "prod-001", SKU: "SKU001", Quantity: 2, UnitPrice: 10.0, TotalPrice: 20.0},
		},
		TotalAmount:     20.0,
		PaymentStatus:   "paid",
		ShippingAddress: ordermanagement.Address{City: "NYC"},
	}
	if err := svc.Create(context.Background(), order); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.Status != ordermanagement.OrderStatusPending {
		t.Errorf("expected pending, got %v", order.Status)
	}
	if err := svc.UpdateStatus(context.Background(), "order-001", ordermanagement.OrderStatusConfirmed); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if err := svc.UpdateStatus(context.Background(), "order-001", ordermanagement.OrderStatusProcessing); err != nil {
		t.Fatalf("process: %v", err)
	}
	if err := svc.UpdateStatus(context.Background(), "order-001", ordermanagement.OrderStatusShipped); err != nil {
		t.Fatalf("ship: %v", err)
	}
	if err := svc.UpdateStatus(context.Background(), "order-001", ordermanagement.OrderStatusDelivered); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	got, _ := svc.GetByID(context.Background(), "order-001")
	if got.Status != ordermanagement.OrderStatusDelivered {
		t.Errorf("expected delivered, got %v", got.Status)
	}
}

func TestOrderInvalidTransition(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	order := &ordermanagement.Order{
		ID: "order-bad", UserID: "u1",
		Items: []ordermanagement.OrderItem{{ID: "i1", ProductID: "p1", Quantity: 1, UnitPrice: 5}},
	}
	if err := svc.Create(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	err := svc.UpdateStatus(context.Background(), "order-bad", ordermanagement.OrderStatusDelivered)
	if err != ordermanagement.ErrInvalidStatusTransition {
		t.Errorf("expected invalid transition, got %v", err)
	}
}

func TestOrderCancel(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	order := &ordermanagement.Order{
		ID: "order-cancel", UserID: "u1",
		Items: []ordermanagement.OrderItem{{ID: "i1", ProductID: "p1", Quantity: 1, UnitPrice: 5}},
	}
	if err := svc.Create(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if err := svc.Cancel(context.Background(), "order-cancel"); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	got, _ := svc.GetByID(context.Background(), "order-cancel")
	if got.Status != ordermanagement.OrderStatusCancelled {
		t.Errorf("expected cancelled, got %v", got.Status)
	}
}

func TestOrderNotFound(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	_, err := svc.GetByID(context.Background(), "nonexistent")
	if err != ordermanagement.ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestOrderListFilter(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	for i := 0; i < 5; i++ {
		order := &ordermanagement.Order{
			ID: "order-flt-" + string(rune('0'+i)), UserID: "u1",
			Items: []ordermanagement.OrderItem{{ID: "i1", ProductID: "p1", Quantity: 1, UnitPrice: 5}},
		}
		svc.Create(context.Background(), order)
	}
	svc.UpdateStatus(context.Background(), "order-flt-0", ordermanagement.OrderStatusConfirmed)
	list, total, err := svc.List(context.Background(), ordermanagement.OrderFilter{Status: ordermanagement.OrderStatusPending})
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 {
		t.Errorf("expected 4 pending, got %d", total)
	}
	_ = list
}

func TestOrderCreateInvalidNoItems(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	err := svc.Create(context.Background(), &ordermanagement.Order{ID: "o1", UserID: "u1"})
	if err != ordermanagement.ErrEmptyItems {
		t.Errorf("expected ErrEmptyItems, got %v", err)
	}
}

func TestOrderStatusTransitions(t *testing.T) {
	tests := []struct {
		from ordermanagement.OrderStatus
		to   ordermanagement.OrderStatus
		ok   bool
	}{
		{ordermanagement.OrderStatusPending, ordermanagement.OrderStatusConfirmed, true},
		{ordermanagement.OrderStatusPending, ordermanagement.OrderStatusCancelled, true},
		{ordermanagement.OrderStatusConfirmed, ordermanagement.OrderStatusProcessing, true},
		{ordermanagement.OrderStatusProcessing, ordermanagement.OrderStatusShipped, true},
		{ordermanagement.OrderStatusShipped, ordermanagement.OrderStatusDelivered, true},
		{ordermanagement.OrderStatusDelivered, ordermanagement.OrderStatusReturned, true},
		{ordermanagement.OrderStatusPending, ordermanagement.OrderStatusDelivered, false},
		{ordermanagement.OrderStatusDelivered, ordermanagement.OrderStatusCancelled, false},
		{ordermanagement.OrderStatusCancelled, ordermanagement.OrderStatusPending, false},
		{ordermanagement.OrderStatusReturned, ordermanagement.OrderStatusPending, false},
	}
	for _, tc := range tests {
		got := ordermanagement.IsValidOrderTransition(tc.from, tc.to)
		if got != tc.ok {
			t.Errorf("IsValidOrderTransition(%v -> %v) = %v, want %v", tc.from, tc.to, got, tc.ok)
		}
	}
}

func TestOrderTimestamps(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	order := &ordermanagement.Order{
		ID: "order-ts", UserID: "u1",
		Items: []ordermanagement.OrderItem{{ID: "i1", ProductID: "p1", Quantity: 1, UnitPrice: 5}},
	}
	if err := svc.Create(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if order.CreatedAt.IsZero() {
		t.Error("created_at should be set")
	}
	if order.UpdatedAt.Before(order.CreatedAt) {
		t.Error("updated_at should not be before created_at")
	}
}

func TestOrderTotalCalculation(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	order := &ordermanagement.Order{
		ID: "order-total", UserID: "u1",
		Items: []ordermanagement.OrderItem{
			{ID: "i1", ProductID: "p1", Quantity: 3, UnitPrice: 10.0},
		},
	}
	order.Items[0].TotalPrice = float64(order.Items[0].Quantity) * order.Items[0].UnitPrice
	if err := svc.Create(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	if order.Items[0].TotalPrice != 30.0 {
		t.Errorf("expected total_price 30.0, got %f", order.Items[0].TotalPrice)
	}
}

func TestInventoryReserveAndRelease(t *testing.T) {
	stockRepo := inventory.NewInMemoryStockRepository()
	svc := inventory.NewService(inventory.NewInMemoryReservationRepository(), stockRepo)

	stockRepo.Upsert(context.Background(), &inventory.Stock{
		ProductID: "prod-001", WarehouseID: "default",
		Available: 100, Reserved: 0, Total: 100,
	})
	res, err := svc.Reserve(context.Background(), inventory.ReserveRequest{
		OrderID: "order-001", ProductID: "prod-001", SKU: "SKU001", Quantity: 10,
	})
	if err != nil {
		t.Fatalf("reserve failed: %v", err)
	}
	if res.Status != inventory.ReservationReserved {
		t.Errorf("expected reserved, got %v", res.Status)
	}
	stock, _ := stockRepo.Get(context.Background(), "prod-001", "default")
	if stock.Available != 90 {
		t.Errorf("expected 90 available, got %d", stock.Available)
	}
	if stock.Reserved != 10 {
		t.Errorf("expected 10 reserved, got %d", stock.Reserved)
	}
	if err := svc.Release(context.Background(), res.ID); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	stock, _ = stockRepo.Get(context.Background(), "prod-001", "default")
	if stock.Available != 100 {
		t.Errorf("expected 100 available after release, got %d", stock.Available)
	}
	if stock.Reserved != 0 {
		t.Errorf("expected 0 reserved after release, got %d", stock.Reserved)
	}
}

func TestInventoryReserveInsufficient(t *testing.T) {
	stockRepo := inventory.NewInMemoryStockRepository()
	svc := inventory.NewService(inventory.NewInMemoryReservationRepository(), stockRepo)

	stockRepo.Upsert(context.Background(), &inventory.Stock{
		ProductID: "prod-001", WarehouseID: "default",
		Available: 5, Reserved: 0, Total: 5,
	})
	_, err := svc.Reserve(context.Background(), inventory.ReserveRequest{
		OrderID: "order-001", ProductID: "prod-001", SKU: "SKU001", Quantity: 10,
	})
	if err != inventory.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestInventoryConsume(t *testing.T) {
	stockRepo := inventory.NewInMemoryStockRepository()
	svc := inventory.NewService(inventory.NewInMemoryReservationRepository(), stockRepo)

	stockRepo.Upsert(context.Background(), &inventory.Stock{
		ProductID: "prod-001", WarehouseID: "default",
		Available: 50, Reserved: 0, Total: 50,
	})
	res, err := svc.Reserve(context.Background(), inventory.ReserveRequest{
		OrderID: "order-001", ProductID: "prod-001", SKU: "SKU001", Quantity: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Consume(context.Background(), res.ID); err != nil {
		t.Fatalf("consume failed: %v", err)
	}
	stock, _ := stockRepo.Get(context.Background(), "prod-001", "default")
	if stock.Available != 40 {
		t.Errorf("expected 40 available, got %d", stock.Available)
	}
	if stock.Reserved != 0 {
		t.Errorf("expected 0 reserved, got %d", stock.Reserved)
	}
	if stock.Total != 40 {
		t.Errorf("expected 40 total, got %d", stock.Total)
	}
}

func TestInventoryCheckAvailability(t *testing.T) {
	stockRepo := inventory.NewInMemoryStockRepository()
	svc := inventory.NewService(inventory.NewInMemoryReservationRepository(), stockRepo)

	stockRepo.Upsert(context.Background(), &inventory.Stock{
		ProductID: "prod-001", WarehouseID: "default",
		Available: 10, Reserved: 0, Total: 10,
	})
	ok, err := svc.CheckAvailability(context.Background(), "prod-001", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected available")
	}
	ok, _ = svc.CheckAvailability(context.Background(), "prod-001", 15)
	if ok {
		t.Error("expected not available")
	}
}

func TestInventoryListStock(t *testing.T) {
	stockRepo := inventory.NewInMemoryStockRepository()
	svc := inventory.NewService(inventory.NewInMemoryReservationRepository(), stockRepo)

	stockRepo.Upsert(context.Background(), &inventory.Stock{ProductID: "p1", WarehouseID: "default", Available: 10, Total: 10})
	stockRepo.Upsert(context.Background(), &inventory.Stock{ProductID: "p2", WarehouseID: "default", Available: 20, Total: 20})
	stocks, err := svc.ListStock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(stocks) != 2 {
		t.Errorf("expected 2 stocks, got %d", len(stocks))
	}
}

func TestPickListCreateAndComplete(t *testing.T) {
	pickRepo := pickpack.NewInMemoryPickListRepository()
	svc := pickpack.NewService(pickRepo, pickpack.NewInMemoryPackingRepository(), pickpack.NewInMemoryShipmentRepository())

	pl := &pickpack.PickList{
		ID: "pick-001", OrderID: "order-001", WarehouseID: "wh-1",
		Items: []pickpack.PickItem{{ProductID: "prod-001", SKU: "SKU001", Quantity: 2}},
	}
	if err := svc.CreatePickList(context.Background(), pl); err != nil {
		t.Fatalf("create pick list: %v", err)
	}
	if pl.Status != pickpack.PickListPending {
		t.Errorf("expected pending, got %v", pl.Status)
	}
	if err := svc.AssignPickList(context.Background(), "pick-001", "worker-1"); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if err := svc.CompletePick(context.Background(), "pick-001"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	got, _ := pickRepo.GetByID(context.Background(), "pick-001")
	if got.Status != pickpack.PickListCompleted {
		t.Errorf("expected completed, got %v", got.Status)
	}
	if got.CompletedAt == nil {
		t.Error("completed_at should be set")
	}
}

func TestPickListNotFound(t *testing.T) {
	svc := pickpack.NewService(pickpack.NewInMemoryPickListRepository(), pickpack.NewInMemoryPackingRepository(), pickpack.NewInMemoryShipmentRepository())
	err := svc.CompletePick(context.Background(), "nonexistent")
	if err != pickpack.ErrPickListNotFound {
		t.Errorf("expected ErrPickListNotFound, got %v", err)
	}
}

func TestPickListInvalidData(t *testing.T) {
	svc := pickpack.NewService(pickpack.NewInMemoryPickListRepository(), pickpack.NewInMemoryPackingRepository(), pickpack.NewInMemoryShipmentRepository())
	err := svc.CreatePickList(context.Background(), &pickpack.PickList{})
	if err != pickpack.ErrInvalidPickData {
		t.Errorf("expected ErrInvalidPickData, got %v", err)
	}
}

func TestPackingCreate(t *testing.T) {
	packRepo := pickpack.NewInMemoryPackingRepository()
	svc := pickpack.NewService(pickpack.NewInMemoryPickListRepository(), packRepo, pickpack.NewInMemoryShipmentRepository())

	p := &pickpack.Packing{
		ID: "pack-001", PickListID: "pick-001", PackageID: "pkg-001",
		Weight: 2.5, Dimensions: "10x10x10",
	}
	if err := svc.CreatePacking(context.Background(), p); err != nil {
		t.Fatalf("create packing: %v", err)
	}
	if p.Status != pickpack.PackingPending {
		t.Errorf("expected pending, got %v", p.Status)
	}
}

func TestShipmentCreate(t *testing.T) {
	shipRepo := pickpack.NewInMemoryShipmentRepository()
	svc := pickpack.NewService(pickpack.NewInMemoryPickListRepository(), pickpack.NewInMemoryPackingRepository(), shipRepo)

	sh := &pickpack.Shipment{
		ID: "ship-001", PackingID: "pack-001",
		Carrier: "DHL", TrackingNumber: "TN123456789",
	}
	if err := svc.CreateShipment(context.Background(), sh); err != nil {
		t.Fatalf("create shipment: %v", err)
	}
	if sh.Status != pickpack.ShipmentPending {
		t.Errorf("expected pending, got %v", sh.Status)
	}
}

func TestPickPackShipWorkflow(t *testing.T) {
	pickRepo := pickpack.NewInMemoryPickListRepository()
	packRepo := pickpack.NewInMemoryPackingRepository()
	shipRepo := pickpack.NewInMemoryShipmentRepository()
	svc := pickpack.NewService(pickRepo, packRepo, shipRepo)

	pl := &pickpack.PickList{
		ID: "pick-wf", OrderID: "order-wf", WarehouseID: "wh-1",
		Items: []pickpack.PickItem{{ProductID: "prod-001", SKU: "SKU001", Quantity: 1}},
	}
	if err := svc.CreatePickList(context.Background(), pl); err != nil {
		t.Fatal(err)
	}
	if err := svc.CompletePick(context.Background(), "pick-wf"); err != nil {
		t.Fatal(err)
	}
	p := &pickpack.Packing{ID: "pack-wf", PickListID: "pick-wf", PackageID: "pkg-wf", Weight: 1.0, Dimensions: "5x5x5"}
	if err := svc.CreatePacking(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	sh := &pickpack.Shipment{ID: "ship-wf", PackingID: "pack-wf", Carrier: "UPS", TrackingNumber: "TRK001"}
	if err := svc.CreateShipment(context.Background(), sh); err != nil {
		t.Fatal(err)
	}
	got, _ := shipRepo.GetByID(context.Background(), "ship-wf")
	if got.Carrier != "UPS" {
		t.Errorf("expected UPS, got %s", got.Carrier)
	}
}

func TestReturnLifecycle(t *testing.T) {
	repo := returns.NewInMemoryRepository()
	svc := returns.NewService(repo)

	ret := &returns.Return{
		ID: "ret-001", OrderID: "order-001", UserID: "user-001",
		Items:  []returns.ReturnItem{{ItemID: "item-001", Quantity: 1, Condition: "damaged"}},
		Reason: "defective item",
	}
	if err := svc.RequestReturn(context.Background(), ret); err != nil {
		t.Fatalf("request return: %v", err)
	}
	if ret.Status != returns.ReturnStatusRequested {
		t.Errorf("expected requested, got %v", ret.Status)
	}
	if ret.RMNumber == "" {
		t.Error("RMA number should be generated")
	}
	if err := svc.ApproveReturn(context.Background(), "ret-001"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if err := svc.ReceiveReturn(context.Background(), "ret-001"); err != nil {
		t.Fatalf("receive: %v", err)
	}
	if err := svc.ProcessRefund(context.Background(), "ret-001", 20.0); err != nil {
		t.Fatalf("refund: %v", err)
	}
	got, _ := repo.GetByID(context.Background(), "ret-001")
	if got.Status != returns.ReturnStatusRefunded {
		t.Errorf("expected refunded, got %v", got.Status)
	}
	if got.RefundAmount != 20.0 {
		t.Errorf("expected refund 20.0, got %f", got.RefundAmount)
	}
}

func TestReturnReject(t *testing.T) {
	repo := returns.NewInMemoryRepository()
	svc := returns.NewService(repo)
	ret := &returns.Return{
		ID: "ret-reject", OrderID: "order-002", UserID: "user-002",
		Items:  []returns.ReturnItem{{ItemID: "item-002", Quantity: 1, Condition: "used"}},
		Reason: "changed mind",
	}
	if err := svc.RequestReturn(context.Background(), ret); err != nil {
		t.Fatal(err)
	}
	if err := svc.RejectReturn(context.Background(), "ret-reject"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	got, _ := repo.GetByID(context.Background(), "ret-reject")
	if got.Status != returns.ReturnStatusRejected {
		t.Errorf("expected rejected, got %v", got.Status)
	}
}

func TestReturnInvalidTransition(t *testing.T) {
	svc := returns.NewService(returns.NewInMemoryRepository())
	ret := &returns.Return{
		ID: "ret-bad", OrderID: "o1", UserID: "u1",
		Items: []returns.ReturnItem{{ItemID: "i1", Quantity: 1}},
	}
	if err := svc.RequestReturn(context.Background(), ret); err != nil {
		t.Fatal(err)
	}
	err := svc.ReceiveReturn(context.Background(), "ret-bad")
	if err != returns.ErrInvalidReturnStatus {
		t.Errorf("expected ErrInvalidReturnStatus, got %v", err)
	}
}

func TestReturnStatusTransitions(t *testing.T) {
	tests := []struct {
		from returns.ReturnStatus
		to   returns.ReturnStatus
		ok   bool
	}{
		{returns.ReturnStatusRequested, returns.ReturnStatusApproved, true},
		{returns.ReturnStatusRequested, returns.ReturnStatusRejected, true},
		{returns.ReturnStatusApproved, returns.ReturnStatusReceived, true},
		{returns.ReturnStatusReceived, returns.ReturnStatusRefunded, true},
		{returns.ReturnStatusRejected, returns.ReturnStatusApproved, false},
		{returns.ReturnStatusRefunded, returns.ReturnStatusRequested, false},
	}
	for _, tc := range tests {
		got := returns.IsValidReturnTransition(tc.from, tc.to)
		if got != tc.ok {
			t.Errorf("IsValidReturnTransition(%v -> %v) = %v, want %v", tc.from, tc.to, got, tc.ok)
		}
	}
}

func TestReturnNotFound(t *testing.T) {
	svc := returns.NewService(returns.NewInMemoryRepository())
	err := svc.ApproveReturn(context.Background(), "nonexistent")
	if err != returns.ErrReturnNotFound {
		t.Errorf("expected ErrReturnNotFound, got %v", err)
	}
}

func TestReturnRefundTimestamps(t *testing.T) {
	repo := returns.NewInMemoryRepository()
	svc := returns.NewService(repo)
	ret := &returns.Return{
		ID: "ret-ts", OrderID: "o1", UserID: "u1",
		Items: []returns.ReturnItem{{ItemID: "i1", Quantity: 1}},
	}
	svc.RequestReturn(context.Background(), ret)
	svc.ApproveReturn(context.Background(), "ret-ts")
	if r, _ := repo.GetByID(context.Background(), "ret-ts"); r.ApprovedAt == nil {
		t.Error("approved_at should be set")
	}
	svc.ReceiveReturn(context.Background(), "ret-ts")
	if r, _ := repo.GetByID(context.Background(), "ret-ts"); r.ReceivedAt == nil {
		t.Error("received_at should be set")
	}
	svc.ProcessRefund(context.Background(), "ret-ts", 15.0)
	if r, _ := repo.GetByID(context.Background(), "ret-ts"); r.RefundedAt == nil {
		t.Error("refunded_at should be set")
	}
}

func TestWarehouseCreate(t *testing.T) {
	whRepo := warehouse.NewInMemoryWarehouseRepository()
	svc := warehouse.NewService(whRepo, warehouse.NewInMemoryZoneRepository(), warehouse.NewInMemoryMovementRepository())

	w := &warehouse.Warehouse{
		ID: "wh-001", Name: "Main Warehouse", City: "NYC", State: "NY",
		IsActive: true, Capacity: 10000,
	}
	if err := svc.CreateWarehouse(context.Background(), w); err != nil {
		t.Fatalf("create warehouse: %v", err)
	}
	got, _ := whRepo.GetByID(context.Background(), "wh-001")
	if got.Name != "Main Warehouse" {
		t.Errorf("expected Main Warehouse, got %s", got.Name)
	}
}

func TestWarehouseList(t *testing.T) {
	whRepo := warehouse.NewInMemoryWarehouseRepository()
	svc := warehouse.NewService(whRepo, warehouse.NewInMemoryZoneRepository(), warehouse.NewInMemoryMovementRepository())

	whRepo.Create(context.Background(), &warehouse.Warehouse{ID: "wh-1", Name: "WH1", IsActive: true})
	whRepo.Create(context.Background(), &warehouse.Warehouse{ID: "wh-2", Name: "WH2", IsActive: true})
	list, err := svc.ListWarehouses(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 warehouses, got %d", len(list))
	}
}

func TestWarehouseZoneCreate(t *testing.T) {
	zoneRepo := warehouse.NewInMemoryZoneRepository()
	svc := warehouse.NewService(warehouse.NewInMemoryWarehouseRepository(), zoneRepo, warehouse.NewInMemoryMovementRepository())

	z := &warehouse.Zone{ID: "zone-001", WarehouseID: "wh-001", Name: "Aisle A", Type: warehouse.ZoneTypeStorage}
	if err := svc.CreateZone(context.Background(), z); err != nil {
		t.Fatalf("create zone: %v", err)
	}
	zones, _ := svc.GetZones(context.Background(), "wh-001")
	if len(zones) != 1 {
		t.Errorf("expected 1 zone, got %d", len(zones))
	}
}

func TestInventoryMovementRecord(t *testing.T) {
	moveRepo := warehouse.NewInMemoryMovementRepository()
	svc := warehouse.NewService(warehouse.NewInMemoryWarehouseRepository(), warehouse.NewInMemoryZoneRepository(), moveRepo)

	m := &warehouse.InventoryMovement{
		ID: "mov-001", ProductID: "prod-001", WarehouseID: "wh-001",
		FromZone: "zone-rec", ToZone: "zone-stor", Quantity: 100,
		Type: warehouse.MovementReceive, Reference: "PO-001",
	}
	if err := svc.RecordMovement(context.Background(), m); err != nil {
		t.Fatalf("record movement: %v", err)
	}
	if m.CreatedAt.IsZero() {
		t.Error("created_at should be set")
	}
	movements, _ := svc.ListMovements(context.Background())
	if len(movements) != 1 {
		t.Errorf("expected 1 movement, got %d", len(movements))
	}
}

func TestWarehouseNotFound(t *testing.T) {
	svc := warehouse.NewService(warehouse.NewInMemoryWarehouseRepository(), warehouse.NewInMemoryZoneRepository(), warehouse.NewInMemoryMovementRepository())
	_, err := svc.GetWarehouse(context.Background(), "nonexistent")
	if err != warehouse.ErrWarehouseNotFound {
		t.Errorf("expected ErrWarehouseNotFound, got %v", err)
	}
}

func TestOrderFullLifecycleWithAllTransitions(t *testing.T) {
	svc := ordermanagement.NewService(ordermanagement.NewInMemoryRepository())
	order := &ordermanagement.Order{
		ID: "order-full", UserID: "u1",
		Items: []ordermanagement.OrderItem{{ID: "i1", ProductID: "p1", Quantity: 1, UnitPrice: 10}},
	}
	if err := svc.Create(context.Background(), order); err != nil {
		t.Fatal(err)
	}
	statuses := []ordermanagement.OrderStatus{
		ordermanagement.OrderStatusConfirmed,
		ordermanagement.OrderStatusProcessing,
		ordermanagement.OrderStatusShipped,
		ordermanagement.OrderStatusDelivered,
	}
	for _, s := range statuses {
		if err := svc.UpdateStatus(context.Background(), "order-full", s); err != nil {
			t.Fatalf("transition to %v: %v", s, err)
		}
	}
	got, _ := svc.GetByID(context.Background(), "order-full")
	if got.Status != ordermanagement.OrderStatusDelivered {
		t.Errorf("expected delivered, got %v", got.Status)
	}
}

func TestReservationReleaseRestoresStock(t *testing.T) {
	stockRepo := inventory.NewInMemoryStockRepository()
	svc := inventory.NewService(inventory.NewInMemoryReservationRepository(), stockRepo)

	stockRepo.Upsert(context.Background(), &inventory.Stock{
		ProductID: "prod-001", WarehouseID: "default", Available: 100, Reserved: 0, Total: 100,
	})
	res, _ := svc.Reserve(context.Background(), inventory.ReserveRequest{
		OrderID: "order-001", ProductID: "prod-001", SKU: "SKU001", Quantity: 30,
	})
	stock, _ := stockRepo.Get(context.Background(), "prod-001", "default")
	if stock.Available != 70 {
		t.Errorf("expected 70, got %d", stock.Available)
	}
	svc.Release(context.Background(), res.ID)
	stock, _ = stockRepo.Get(context.Background(), "prod-001", "default")
	if stock.Available != 100 {
		t.Errorf("expected 100 after release, got %d", stock.Available)
	}
}

func TestReservationExpiryTime(t *testing.T) {
	stockRepo := inventory.NewInMemoryStockRepository()
	svc := inventory.NewService(inventory.NewInMemoryReservationRepository(), stockRepo)

	stockRepo.Upsert(context.Background(), &inventory.Stock{
		ProductID: "prod-001", WarehouseID: "default", Available: 10, Reserved: 0, Total: 10,
	})
	res, err := svc.Reserve(context.Background(), inventory.ReserveRequest{
		OrderID: "order-exp", ProductID: "prod-001", SKU: "SKU001", Quantity: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExpiresAt.Before(time.Now().Add(23 * time.Hour)) {
		t.Error("expires_at should be ~24h in the future")
	}
}

func TestPickListAssignUpdatesStatus(t *testing.T) {
	pickRepo := pickpack.NewInMemoryPickListRepository()
	svc := pickpack.NewService(pickRepo, pickpack.NewInMemoryPackingRepository(), pickpack.NewInMemoryShipmentRepository())

	pl := &pickpack.PickList{ID: "pick-assign", OrderID: "order-assign", WarehouseID: "wh-1",
		Items: []pickpack.PickItem{{ProductID: "p1", SKU: "S1", Quantity: 1}},
	}
	svc.CreatePickList(context.Background(), pl)
	if err := svc.AssignPickList(context.Background(), "pick-assign", "worker-1"); err != nil {
		t.Fatal(err)
	}
	got, _ := pickRepo.GetByID(context.Background(), "pick-assign")
	if got.AssignedTo != "worker-1" {
		t.Errorf("expected worker-1, got %s", got.AssignedTo)
	}
	if got.Status != pickpack.PickListInProgress {
		t.Errorf("expected in_progress, got %v", got.Status)
	}
}

func TestWarehouseUpdateUtilization(t *testing.T) {
	whRepo := warehouse.NewInMemoryWarehouseRepository()
	svc := warehouse.NewService(whRepo, warehouse.NewInMemoryZoneRepository(), warehouse.NewInMemoryMovementRepository())

	whRepo.Create(context.Background(), &warehouse.Warehouse{
		ID: "wh-util", Name: "Util WH", IsActive: true, Capacity: 1000, CurrentUtilization: 0,
	})
	w, _ := svc.GetWarehouse(context.Background(), "wh-util")
	w.CurrentUtilization = 500
	whRepo.Update(context.Background(), w)
	got, _ := whRepo.GetByID(context.Background(), "wh-util")
	if got.CurrentUtilization != 500 {
		t.Errorf("expected 500, got %d", got.CurrentUtilization)
	}
}
