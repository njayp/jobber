package server

import (
	"context"

	"github.com/njayp/jobber/pkg/pb"
)

func (s *Service) Start(_ context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	return s.manager.Start(req)
}

func (s *Service) Stop(_ context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	return s.manager.Stop(req)
}

func (s *Service) Status(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	return s.manager.Status(req)
}

type StreamWriter struct {
	send func(*pb.StreamResponse) error
}

func (s *StreamWriter) Write(p []byte) (int, error) {
	err := s.send(&pb.StreamResponse{Data: p})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *Service) Stream(req *pb.StreamRequest, stream pb.Jobber_StreamServer) error {
	return s.manager.Stream(stream.Context(), req, &StreamWriter{send: stream.Send})
}

func (s *Service) Version(ctx context.Context, req *pb.VersionRequest) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{Version: "1.0.0-alpha"}, nil
}
