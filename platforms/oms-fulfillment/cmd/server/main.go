package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/inventory"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/ordermanagement"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/pickpack"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/returns"
	httptransport "github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/transport/http"
	"github.com/tikiclone/tiki/platforms/oms-fulfillment/internal/warehouse"
	automaxprocs "go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// Tune GC for low-latency: more frequent GCs, less heap growth
	if gogc := os.Getenv("GOGC"); gogc == "" {
		os.Setenv("GOGC", "50")
	}
}

func main() {
	orderRepo := ordermanagement.NewInMemoryRepository()
	// Auto-tune GOMAXPROCS for container environments
	_, _ = automaxprocs.Set()

	inventoryResRepo := inventory.NewInMemoryReservationRepository()
	stockRepo := inventory.NewInMemoryStockRepository()
	pickListRepo := pickpack.NewInMemoryPickListRepository()
	packingRepo := pickpack.NewInMemoryPackingRepository()
	shipmentRepo := pickpack.NewInMemoryShipmentRepository()
	returnRepo := returns.NewInMemoryRepository()
	warehouseRepo := warehouse.NewInMemoryWarehouseRepository()
	zoneRepo := warehouse.NewInMemoryZoneRepository()
	movementRepo := warehouse.NewInMemoryMovementRepository()

	orderSvc := ordermanagement.NewService(orderRepo)
	inventorySvc := inventory.NewService(inventoryResRepo, stockRepo)
	pickpackSvc := pickpack.NewService(pickListRepo, packingRepo, shipmentRepo)
	returnSvc := returns.NewService(returnRepo)
	warehouseSvc := warehouse.NewService(warehouseRepo, zoneRepo, movementRepo)

	handler := httptransport.NewHandler(orderSvc, inventorySvc, pickpackSvc, returnSvc, warehouseSvc)
	router := httptransport.SetupRouter(handler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:       5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("starting oms-fulfillment server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

