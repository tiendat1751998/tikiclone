package grpc

import (
	"context"

	"github.com/tikiclone/tiki/packages/go-shared/pkg/errors"
	"github.com/tikiclone/tiki/services/catalog-product/internal/domain"
	"github.com/tikiclone/tiki/services/catalog-product/internal/usecase"
	pb "github.com/tikiclone/tiki/proto/catalog/v1"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CatalogGRPCServer struct {
	pb.UnimplementedCatalogServiceServer
	productUseCase  *usecase.ProductUseCase
	categoryUseCase *usecase.CategoryUseCase
}

func NewCatalogGRPCServer(productUC *usecase.ProductUseCase, categoryUC *usecase.CategoryUseCase) *CatalogGRPCServer {
	return &CatalogGRPCServer{
		productUseCase:  productUC,
		categoryUseCase: categoryUC,
	}
}

func (s *CatalogGRPCServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	ctx, span := otel.Tracer("catalog-product").Start(ctx, "grpc.product.get")
	defer span.End()

	product, err := s.productUseCase.GetByID(ctx, req.GetSpuId())
	if err != nil {
		appErr := errors.FromError(err)
		switch appErr.Code {
		case errors.ErrNotFound:
			return nil, status.Error(codes.NotFound, appErr.Message)
		default:
			return nil, status.Error(codes.Internal, appErr.Message)
		}
	}

	return &pb.GetProductResponse{Product: domainToProtoProduct(product)}, nil
}

func (s *CatalogGRPCServer) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *CatalogGRPCServer) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *CatalogGRPCServer) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *CatalogGRPCServer) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func domainToProtoProduct(p *domain.Product) *pb.Product {
	skus := make([]*pb.SKU, len(p.SKUs))
	for i, s := range p.SKUs {
		vars := make([]*pb.Variation, len(s.Variations))
		for j, v := range s.Variations {
			vars[j] = &pb.Variation{Name: v.Name, Value: v.Value}
		}
		skus[i] = &pb.SKU{
			SkuId:      s.SKUID,
			SpuId:      s.SPUID,
			Price:      s.Price,
			Stock:      s.Stock,
			Variations: vars,
			Image:      s.Image,
			Status:     s.Status,
		}
	}

	attrs := make(map[string]string)
	if p.Attributes != nil {
		attrs = p.Attributes
	}

	return &pb.Product{
		SpuId:       p.SPUID,
		Title:       p.Title,
		Description: p.Description,
		CategoryId:  p.CategoryID,
		Skus:        skus,
		Attributes:  attrs,
		Images:      p.Images,
		SellerId:    p.SellerID,
		Status:      p.Status,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}
