package server

import (
	"context"
	"io"
	"net"

	"github.com/njayp/jobber/pkg/pb"
	"github.com/njayp/jobber/pkg/server/authn"
	"github.com/njayp/jobber/pkg/server/authz"
	"google.golang.org/grpc"
)

type Manager interface {
	Start(*pb.StartRequest) (*pb.StartResponse, error)
	Stop(*pb.StopRequest) (*pb.StopResponse, error)
	Status(*pb.StatusRequest) (*pb.StatusResponse, error)
	Stream(context.Context, *pb.StreamRequest, io.Writer) error
}

type Service struct {
	pb.UnimplementedJobberServer
	manager Manager
}

func (s *Service) Serve(url string) error {
	lis, err := net.Listen("tcp", url)
	if err != nil {
		return err
	}

	srv, err := newAuthServer()
	if err != nil {
		return err
	}

	pb.RegisterJobberServer(srv, s)
	return srv.Serve(lis)
}

func newAuthServer() (*grpc.Server, error) {
	creds, err := authn.LoadTLSCredentials()
	if err != nil {
		return nil, err
	}

	rbac := authz.NewRBAC()

	return grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(rbac.UnaryInterceptor),
		grpc.StreamInterceptor(rbac.StreamInterceptor),
	), nil
}
