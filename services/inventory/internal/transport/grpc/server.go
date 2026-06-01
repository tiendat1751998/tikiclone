package grpc

import (
	"context"

	"github.com/tikiclone/tiki/services/inventory/internal/application"
	"github.com/tikiclone/tiki/services/inventory/internal/domain"
	pb "github.com/tikiclone/tiki/services/inventory/proto/inventory/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryGRPCServer struct {
	pb.UnimplementedInventoryServiceServer
	inventoryService *application.InventoryService
}

func NewInventoryGRPCServer(svc *application.InventoryService) *InventoryGRPCServer {
	return &InventoryGRPCServer{inventoryService: svc}
}

func (s *InventoryGRPCServer) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReservationResponse, error) {
	r := &application.ReserveStockRequest{
		OrderID: req.OrderId, UserID: req.UserId, ProductID: req.ProductId,
		SkuID: req.SkuId, WarehouseID: req.WarehouseId, Quantity: int(req.Quantity),
		IdempotencyKey: req.IdempotencyKey,
	}
	res, err := s.inventoryService.ReserveStock(ctx, r)
	if err != nil { return nil, toGRPCError(err) }
	return &pb.ReservationResponse{Id: res.ID, OrderId: res.OrderID, SkuId: res.SkuID, Quantity: int32(res.Quantity), Status: string(res.Status)}, nil
}

func (s *InventoryGRPCServer) ReleaseStock(ctx context.Context, req *pb.ReleaseStockRequest) (*pb.ReleaseStockResponse, error) {
	if err := s.inventoryService.ReleaseStock(ctx, req.ReservationId); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ReleaseStockResponse{Status: "released"}, nil
}

func (s *InventoryGRPCServer) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.StockResponse, error) {
	stock, err := s.inventoryService.GetStock(ctx, req.SkuId, req.WarehouseId)
	if err != nil { return nil, toGRPCError(err) }
	return &pb.StockResponse{Id: stock.ID, SkuId: stock.SkuID, Quantity: int32(stock.Quantity), ReservedQty: int32(stock.ReservedQty), AvailableQty: int32(stock.AvailableQty), Status: string(stock.Status)}, nil
}

func toGRPCError(err error) error {
	switch err {
	case domain.ErrStockNotFound: return status.Error(codes.NotFound, err.Error())
	case domain.ErrInsufficientStock: return status.Error(codes.FailedPrecondition, err.Error())
	case domain.ErrReservationNotFound: return status.Error(codes.NotFound, err.Error())
	default: zap.L().Error("unexpected gRPC error", zap.Error(err)); return status.Error(codes.Internal, "internal server error")
	}
}
