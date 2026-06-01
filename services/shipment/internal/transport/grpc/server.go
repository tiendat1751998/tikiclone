package grpc

import (
	"context"

	"github.com/tikiclone/tiki/services/shipment/internal/application"
	"github.com/tikiclone/tiki/services/shipment/internal/domain"
	pb "github.com/tikiclone/tiki/services/shipment/proto/shipment/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShipmentGRPCServer struct {
	pb.UnimplementedShipmentServiceServer
	shipmentService *application.ShipmentService
}

func NewShipmentGRPCServer(svc *application.ShipmentService) *ShipmentGRPCServer {
	return &ShipmentGRPCServer{shipmentService: svc}
}

func (s *ShipmentGRPCServer) CreateShipment(ctx context.Context, req *pb.CreateShipmentRequest) (*pb.ShipmentResponse, error) {
	r := &application.CreateShipmentRequest{
		OrderID: req.OrderId, UserID: req.UserId, CarrierID: req.CarrierId,
		IdempotencyKey: req.IdempotencyKey, Weight: req.Weight, Currency: req.Currency,
	}
	shipment, err := s.shipmentService.CreateShipment(ctx, r)
	if err != nil { return nil, toGRPCError(err) }
	return toShipmentResponse(shipment), nil
}

func (s *ShipmentGRPCServer) GetShipment(ctx context.Context, req *pb.GetShipmentRequest) (*pb.ShipmentResponse, error) {
	shipment, err := s.shipmentService.GetShipment(ctx, req.ShipmentId)
	if err != nil { return nil, toGRPCError(err) }
	return toShipmentResponse(shipment), nil
}

func (s *ShipmentGRPCServer) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.ShipmentResponse, error) {
	shipment, err := s.shipmentService.UpdateStatus(ctx, req.ShipmentId, domain.ShipmentStatus(req.Status), req.ActorId, req.Reason)
	if err != nil { return nil, toGRPCError(err) }
	return toShipmentResponse(shipment), nil
}

func toShipmentResponse(s *domain.Shipment) *pb.ShipmentResponse {
	return &pb.ShipmentResponse{
		Id: s.ID, OrderId: s.OrderID, UserId: s.UserID, CarrierId: s.CarrierID,
		TrackingNumber: s.TrackingNumber, Status: string(s.Status),
		Cost: s.Cost, Currency: s.Currency, CreatedAt: s.CreatedAt.String(),
	}
}

func toGRPCError(err error) error {
	switch err {
	case domain.ErrShipmentNotFound: return status.Error(codes.NotFound, err.Error())
	case domain.ErrInvalidShipmentState: return status.Error(codes.FailedPrecondition, err.Error())
	default: zap.L().Error("unexpected gRPC error", zap.Error(err)); return status.Error(codes.Internal, "internal server error")
	}
}
