/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package api

import (
	"context"

	proxy "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type RestProxyService interface {
	RegisterWithRestProxy(context.Context, *proxy.ServeMux, []grpc.DialOption, string) error
}

type Service interface {
	Start() error
	Stop() error
	Name() string
}
