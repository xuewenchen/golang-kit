package rpc

import (
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"golang-kit/config"
	"golang-kit/net/etcd"
	"google.golang.org/grpc"
	"time"
)

type GrpcClient struct {
	Conn *grpc.ClientConn
}

func NewClient(c *config.GrpcClient, tracer opentracing.Tracer) (client *GrpcClient, err error) {
	var conn *grpc.ClientConn
	r := etcd.NewResolver(c.ServiceName)
	b := grpc.RoundRobin(r)
	conn, err = grpc.Dial(
		c.RegisterAddr,
		grpc.WithInsecure(),
		grpc.WithBalancer(b),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer, otgrpc.LogPayloads())),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*2),
	)

	client = &GrpcClient{
		Conn: conn,
	}
	return
}
