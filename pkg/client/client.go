package client

import (
	"github.com/njayp/jobber/pkg/pb"
	"google.golang.org/grpc"
)

func NewJobberClient(url string) (pb.JobberClient, error) {
	creds, err := loadClientTLSCredentials()
	if err != nil {
		return nil, err
	}

	cli, err := grpc.NewClient(url, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	return pb.NewJobberClient(cli), nil
}
