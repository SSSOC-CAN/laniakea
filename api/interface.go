package api

import (
	"context"
	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type RestProxyService interface {
	RegisterWithRestProxy(context.Context, *proxy.ServeMux, []grpc.DialOption, string) error
}