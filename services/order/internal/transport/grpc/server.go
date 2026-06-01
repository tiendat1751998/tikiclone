package grpc

import (
	"context"

	"github.com/tikiclone/tiki/services/order/internal/application"
	"github.com/tikiclone/tiki/services/order/internal/domain"
	pb "github.com/tikiclone/tiki/services/order/proto/order/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	orderService *application.OrderService
}

func NewOrderGRPCServer(orderService *application.OrderService) *OrderGRPCServer {
	return &OrderGRPCServer{orderService: orderService}
}

func (s *OrderGRPCServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	items := make([]domain.SnapshotItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.SnapshotItem{
			ProductID: item.ProductId,
			SkuID:     item.SkuId,
			ShopID:    item.ShopId,
			Name:      item.Name,
			Quantity:  int(item.Quantity),
			UnitPrice: item.UnitPrice,
			ImageURL:  item.ImageUrl,
		})
	}
	createReq := &application.CreateOrderRequest{
		UserID:         req.UserId,
		SellerID:       req.SellerId,
		Currency:       req.Currency,
		IdempotencyKey: req.IdempotencyKey,
		ShippingAddress: domain.Address{
			Street1:    req.ShippingAddress.Street1,
			Street2:    req.ShippingAddress.Street2,
			City:       req.ShippingAddress.City,
			State:      req.ShippingAddress.State,
			PostalCode: req.ShippingAddress.PostalCode,
			Country:    req.ShippingAddress.Country,
			Phone:      req.ShippingAddress.Phone,
		},
		Items: items,
	}
	order, err := s.orderService.CreateOrder(ctx, createReq)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toOrderResponse(order), nil
}

func (s *OrderGRPCServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	order, err := s.orderService.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toOrderResponse(order), nil
}

func (s *OrderGRPCServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, total, err := s.orderService.ListOrders(ctx, req.UserId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, toGRPCError(err)
	}
	pbOrders := make([]*pb.OrderResponse, 0, len(orders))
	for _, o := range orders {
		pbOrders = append(pbOrders, toOrderResponse(o))
	}
	return &pb.ListOrdersResponse{Orders: pbOrders, Total: int32(total), Page: req.Page}, nil
}

func (s *OrderGRPCServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	cancelReq := &application.CancelOrderRequest{
		OrderID:       req.OrderId,
		Reason:        req.Reason,
		CancelledBy:   req.CancelledBy,
		CancelledType: domain.CancellationTypeUser,
	}
	order, err := s.orderService.CancelOrder(ctx, cancelReq)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toOrderResponse(order), nil
}

func (s *OrderGRPCServer) UpdateOrderStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.OrderResponse, error) {
	order, err := s.orderService.TransitionStatus(ctx, req.OrderId, domain.OrderStatus(req.Status), req.ActorId, req.ActorType, req.Reason)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toOrderResponse(order), nil
}

func (s *OrderGRPCServer) GetOrderHistory(ctx context.Context, req *pb.GetOrderHistoryRequest) (*pb.OrderHistoryResponse, error) {
	events, err := s.orderService.GetOrderHistory(ctx, req.OrderId)
	if err != nil {
		return nil, toGRPCError(err)
	}
	pbEvents := make([]*pb.LifecycleEvent, 0, len(events))
	for _, e := range events {
		pbEvents = append(pbEvents, &pb.LifecycleEvent{
			Id:               e.ID,
			FromStatus:       string(e.FromStatus),
			ToStatus:         string(e.ToStatus),
			TransitionReason: e.TransitionReason,
			ActorId:          e.ActorID,
			ActorType:        e.ActorType,
			CreatedAt:        e.CreatedAt.String(),
		})
	}
	return &pb.OrderHistoryResponse{OrderId: req.OrderId, Events: pbEvents}, nil
}

func toOrderResponse(order *domain.Order) *pb.OrderResponse {
	items := make([]*pb.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &pb.OrderItem{
			Id:         item.ID,
			ProductId:  item.ProductID,
			SkuId:      item.SkuID,
			ShopId:     item.ShopID,
			Quantity:   int32(item.Quantity),
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		})
	}
	return &pb.OrderResponse{
		Id:          order.ID,
		OrderNumber: order.OrderNumber,
		UserId:      order.UserID,
		SellerId:    order.SellerID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		Currency:    order.Currency,
		ShippingAddress: &pb.Address{
			Street1:    order.ShippingAddress.Street1,
			City:       order.ShippingAddress.City,
			State:      order.ShippingAddress.State,
			PostalCode: order.ShippingAddress.PostalCode,
			Country:    order.ShippingAddress.Country,
		},
		Items:     items,
		Version:   int32(order.Version),
		CreatedAt: order.CreatedAt.String(),
		UpdatedAt: order.UpdatedAt.String(),
	}
}

func toGRPCError(err error) error {
	switch err {
	case domain.ErrOrderNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrOrderNotCancellable:
		return status.Error(codes.FailedPrecondition, err.Error())
	case domain.ErrInvalidStateTransition:
		return status.Error(codes.FailedPrecondition, err.Error())
	case domain.ErrUnauthorized:
		return status.Error(codes.PermissionDenied, err.Error())
	case domain.ErrConcurrentModification:
		return status.Error(codes.Aborted, err.Error())
	default:
		zap.L().Error("unexpected gRPC error", zap.Error(err))
		return status.Error(codes.Internal, "internal server error")
	}
}
