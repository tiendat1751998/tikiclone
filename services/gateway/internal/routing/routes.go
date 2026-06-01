package routing

import "github.com/tikiclone/tiki/services/gateway/internal/transport"

type RouteGroup struct {
	Prefix      string
	Target      string
	Strip       string
	Auth        bool
	RateLimit   int
	Roles       []string
	Protocol    string
	GRPCMethod  string
}

var RouteTable = []RouteGroup{
	{
		Prefix:    "/api/v1/auth",
		Target:    "auth",
		Strip:     "",
		Auth:      false,
		RateLimit: 5000,
	},
	{
		Prefix:    "/api/v1/products",
		Target:    "catalog",
		Strip:     "",
		Auth:      false,
		RateLimit: 10000,
	},
	{
		Prefix:    "/api/v1/categories",
		Target:    "catalog",
		Strip:     "",
		Auth:      false,
		RateLimit: 10000,
	},
	{
		Prefix:    "/api/v1/cart",
		Target:    "cart",
		Strip:     "/api/v1/cart",
		Auth:      true,
		RateLimit: 5000,
	},
	{
		Prefix:    "/api/v1/orders",
		Target:    "order",
		Strip:     "/api/v1/orders",
		Auth:      true,
		RateLimit: 5000,
	},
	{
		Prefix:    "/api/v1/checkout",
		Target:    "order",
		Strip:     "/api/v1/checkout",
		Auth:      true,
		RateLimit: 1000,
	},
	{
		Prefix:    "/api/v1/inventory",
		Target:    "inventory",
		Strip:     "/api/v1/inventory",
		Auth:      true,
		RateLimit: 5000,
		Roles:     []string{"admin", "service"},
	},
	{
		Prefix:    "/api/v1/payments",
		Target:    "payment",
		Strip:     "/api/v1/payments",
		Auth:      true,
		RateLimit: 1000,
	},
	{
		Prefix:    "/api/v1/recommendations",
		Target:    "recommendation",
		Strip:     "/api/v1/recommendations",
		Auth:      true,
		RateLimit: 5000,
	},
	{
		Prefix:    "/api/v1/delivery",
		Target:    "delivery",
		Strip:     "/api/v1/delivery",
		Auth:      false,
		RateLimit: 5000,
	},
	{
		Prefix:    "/api/v1/grpc/inventory",
		Target:    "inventory",
		Strip:     "/api/v1/grpc/inventory",
		Auth:      true,
		RateLimit: 200,
		Roles:     []string{"service"},
		Protocol:  "grpc",
		GRPCMethod: "/tiki.inventory.InventoryService/ReserveStock",
	},
}

func (r *RouteGroup) ToProxyTarget() *transport.ProxyTarget {
	return &transport.ProxyTarget{
		ServiceName: r.Target,
		PathPrefix:  r.Prefix,
		StripPrefix: r.Strip,
	}
}
