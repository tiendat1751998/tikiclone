package grpc

import (
	"context"
	"time"

	"github.com/tikiclone/tiki/services/auth/internal/application"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	pb "github.com/tikiclone/tiki/services/auth/proto/auth/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCServer struct {
	pb.UnimplementedAuthServiceServer
	authService *application.AuthService
}

func NewAuthGRPCServer(authService *application.AuthService) *AuthGRPCServer {
	return &AuthGRPCServer{authService: authService}
}

func (s *AuthGRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.register")
	defer span.End()

	registerReq := &domain.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		Username:  req.Username,
		DisplayName: req.DisplayName,
		Phone:     req.Phone,
	}

	ip := extractIP(ctx)
	userAgent := extractUserAgent(ctx)

	tokens, session, err := s.authService.Register(ctx, registerReq, ip, userAgent)
	if err != nil {
		return nil, mapError(err)
	}

	span.SetAttributes(attribute.String("user_id", tokens.SessionID))

	_ = session

	return &pb.RegisterResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
		SessionId:    tokens.SessionID,
	}, nil
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.login")
	defer span.End()

	loginReq := &domain.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		DeviceID: req.DeviceId,
	}

	ip := extractIP(ctx)
	userAgent := extractUserAgent(ctx)

	tokens, session, err := s.authService.Login(ctx, loginReq, ip, userAgent)
	if err != nil {
		return nil, mapError(err)
	}

	_ = session

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
		SessionId:    tokens.SessionID,
	}, nil
}

func (s *AuthGRPCServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.refresh")
	defer span.End()

	ip := extractIP(ctx)
	userAgent := extractUserAgent(ctx)

	tokens, session, err := s.authService.RefreshToken(ctx, req.RefreshToken, ip, userAgent)
	if err != nil {
		return nil, mapError(err)
	}

	_ = session

	return &pb.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
		SessionId:    tokens.SessionID,
	}, nil
}

func (s *AuthGRPCServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.logout")
	defer span.End()

	err := s.authService.Logout(ctx, req.AccessToken, req.RefreshToken, req.AllDevices)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.LogoutResponse{
		Success: true,
		Message: "logged out successfully",
	}, nil
}

func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.validate")
	defer span.End()

	claims, err := s.authService.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	roles := make([]string, len(claims.Roles))
	for i, r := range claims.Roles {
		roles[i] = string(r)
	}

	return &pb.ValidateTokenResponse{
		Valid:     true,
		UserId:    claims.UserID,
		Email:     claims.Email,
		Roles:     roles,
		SessionId: claims.SessionID,
	}, nil
}

func (s *AuthGRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.get_user")
	defer span.End()

	user, err := s.authService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetUserResponse{
		Id:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		DisplayName:   user.DisplayName,
		Status:        string(user.Status),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *AuthGRPCServer) GetSessions(ctx context.Context, req *pb.GetSessionsRequest) (*pb.GetSessionsResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.sessions")
	defer span.End()

	sessions, err := s.authService.GetActiveSessions(ctx, req.UserId)
	if err != nil {
		return nil, mapError(err)
	}

	pbSessions := make([]*pb.SessionInfo, len(sessions))
	for i, session := range sessions {
		pbSessions[i] = &pb.SessionInfo{
			Id:           session.ID,
			DeviceId:     session.DeviceID,
			DeviceName:   session.DeviceName,
			DeviceType:   session.DeviceType,
			Ip:           session.IP,
			Status:       string(session.Status),
			LastActiveAt: session.LastActiveAt.Format(time.RFC3339),
			CreatedAt:    session.CreatedAt.Format(time.RFC3339),
		}
	}

	return &pb.GetSessionsResponse{
		Sessions: pbSessions,
		Count:    int32(len(pbSessions)),
	}, nil
}

func (s *AuthGRPCServer) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.RevokeSessionResponse, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "grpc.revoke_session")
	defer span.End()

	err := s.authService.RevokeSession(ctx, req.UserId, req.SessionId)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.RevokeSessionResponse{Success: true}, nil
}

func mapError(err error) error {
	switch err {
	case domain.ErrInvalidCredentials:
		return status.Error(grpccodes.Unauthenticated, err.Error())
	case domain.ErrInvalidToken, domain.ErrTokenExpired, domain.ErrInvalidRefreshToken:
		return status.Error(grpccodes.Unauthenticated, err.Error())
	case domain.ErrAccountLocked, domain.ErrRateLimited:
		return status.Error(grpccodes.ResourceExhausted, err.Error())
	case domain.ErrAccountInactive, domain.ErrAccountSuspended, domain.ErrEmailNotVerified:
		return status.Error(grpccodes.PermissionDenied, err.Error())
	case domain.ErrEmailAlreadyExists, domain.ErrUsernameTaken, domain.ErrMaxSessions:
		return status.Error(grpccodes.AlreadyExists, err.Error())
	case domain.ErrPasswordTooWeak:
		return status.Error(grpccodes.InvalidArgument, err.Error())
	case domain.ErrUserNotFound:
		return status.Error(grpccodes.NotFound, err.Error())
	case domain.ErrInsufficientPerms:
		return status.Error(grpccodes.PermissionDenied, err.Error())
	default:
		return status.Error(grpccodes.Internal, "internal server error")
	}
}

func extractIP(ctx context.Context) string {
	return ""
}

func extractUserAgent(ctx context.Context) string {
	return ""
}
