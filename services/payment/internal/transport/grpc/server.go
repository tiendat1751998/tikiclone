package grpc

import (
	"context"

	"github.com/tikiclone/tiki/services/payment/internal/application"
	"github.com/tikiclone/tiki/services/payment/internal/domain"
	pb "github.com/tikiclone/tiki/services/payment/proto/payment/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	paymentService *application.PaymentService
}

func NewPaymentGRPCServer(svc *application.PaymentService) *PaymentGRPCServer {
	return &PaymentGRPCServer{paymentService: svc}
}

func (s *PaymentGRPCServer) AuthorizePayment(ctx context.Context, req *pb.AuthorizePaymentRequest) (*pb.PaymentResponse, error) {
	r := &application.AuthorizePaymentRequest{
		OrderID: req.OrderId, UserID: req.UserId, Amount: req.Amount,
		Currency: req.Currency, PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		IdempotencyKey: req.IdempotencyKey,
	}
	p, err := s.paymentService.AuthorizePayment(ctx, r)
	if err != nil { return nil, toGRPCError(err) }
	return toPaymentResponse(p), nil
}

func (s *PaymentGRPCServer) CapturePayment(ctx context.Context, req *pb.CapturePaymentRequest) (*pb.PaymentResponse, error) {
	p, err := s.paymentService.CapturePayment(ctx, req.PaymentId, req.ActorId)
	if err != nil { return nil, toGRPCError(err) }
	return toPaymentResponse(p), nil
}

func (s *PaymentGRPCServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.PaymentResponse, error) {
	p, err := s.paymentService.GetPayment(ctx, req.PaymentId)
	if err != nil { return nil, toGRPCError(err) }
	return toPaymentResponse(p), nil
}

func toPaymentResponse(p *domain.Payment) *pb.PaymentResponse {
	return &pb.PaymentResponse{
		Id: p.ID, OrderId: p.OrderID, UserId: p.UserID,
		Amount: p.Amount, Currency: p.Currency, Status: string(p.Status),
		PaymentMethod: string(p.PaymentMethod), PspTransactionId: p.PSPTransactionID,
		AmountRefunded: p.AmountRefunded, CreatedAt: p.CreatedAt.String(),
	}
}

func toGRPCError(err error) error {
	switch err {
	case domain.ErrPaymentNotFound: return status.Error(codes.NotFound, err.Error())
	case domain.ErrDoubleChargeDetected: return status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrInvalidWebhookSignature: return status.Error(codes.Unauthenticated, err.Error())
	case domain.ErrRefundNotAllowed: return status.Error(codes.FailedPrecondition, err.Error())
	default:
		zap.L().Error("unexpected gRPC error", zap.Error(err))
		return status.Error(codes.Internal, "internal server error")
	}
}
