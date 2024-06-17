// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.3
// source: jobber.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Jobber_Start_FullMethodName  = "/pb.Jobber/Start"
	Jobber_Stop_FullMethodName   = "/pb.Jobber/Stop"
	Jobber_Status_FullMethodName = "/pb.Jobber/Status"
	Jobber_Stream_FullMethodName = "/pb.Jobber/Stream"
)

// JobberClient is the client API for Jobber service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type JobberClient interface {
	// Start throws error if process does not start.
	Start(ctx context.Context, in *StartRequest, opts ...grpc.CallOption) (*StartResponse, error)
	// Stop does not wait for the cgroup to exit. Status should be used
	// to check whether a process has exited.
	Stop(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*StopResponse, error)
	// IDEA watching functionality should be added to this rpc.
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	// Stream copies and follows one file for neatness and control.
	// StreamSelect selects between "stdout.txt" and "stderr.txt".
	Stream(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (Jobber_StreamClient, error)
}

type jobberClient struct {
	cc grpc.ClientConnInterface
}

func NewJobberClient(cc grpc.ClientConnInterface) JobberClient {
	return &jobberClient{cc}
}

func (c *jobberClient) Start(ctx context.Context, in *StartRequest, opts ...grpc.CallOption) (*StartResponse, error) {
	out := new(StartResponse)
	err := c.cc.Invoke(ctx, Jobber_Start_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobberClient) Stop(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*StopResponse, error) {
	out := new(StopResponse)
	err := c.cc.Invoke(ctx, Jobber_Stop_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobberClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, Jobber_Status_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jobberClient) Stream(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (Jobber_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Jobber_ServiceDesc.Streams[0], Jobber_Stream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &jobberStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Jobber_StreamClient interface {
	Recv() (*StreamResponse, error)
	grpc.ClientStream
}

type jobberStreamClient struct {
	grpc.ClientStream
}

func (x *jobberStreamClient) Recv() (*StreamResponse, error) {
	m := new(StreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// JobberServer is the server API for Jobber service.
// All implementations must embed UnimplementedJobberServer
// for forward compatibility
type JobberServer interface {
	// Start throws error if process does not start.
	Start(context.Context, *StartRequest) (*StartResponse, error)
	// Stop does not wait for the cgroup to exit. Status should be used
	// to check whether a process has exited.
	Stop(context.Context, *StopRequest) (*StopResponse, error)
	// IDEA watching functionality should be added to this rpc.
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
	// Stream copies and follows one file for neatness and control.
	// StreamSelect selects between "stdout.txt" and "stderr.txt".
	Stream(*StreamRequest, Jobber_StreamServer) error
	mustEmbedUnimplementedJobberServer()
}

// UnimplementedJobberServer must be embedded to have forward compatible implementations.
type UnimplementedJobberServer struct {
}

func (UnimplementedJobberServer) Start(context.Context, *StartRequest) (*StartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedJobberServer) Stop(context.Context, *StopRequest) (*StopResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedJobberServer) Status(context.Context, *StatusRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (UnimplementedJobberServer) Stream(*StreamRequest, Jobber_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}
func (UnimplementedJobberServer) mustEmbedUnimplementedJobberServer() {}

// UnsafeJobberServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to JobberServer will
// result in compilation errors.
type UnsafeJobberServer interface {
	mustEmbedUnimplementedJobberServer()
}

func RegisterJobberServer(s grpc.ServiceRegistrar, srv JobberServer) {
	s.RegisterService(&Jobber_ServiceDesc, srv)
}

func _Jobber_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobberServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Jobber_Start_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobberServer).Start(ctx, req.(*StartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Jobber_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobberServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Jobber_Stop_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobberServer).Stop(ctx, req.(*StopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Jobber_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JobberServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Jobber_Status_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JobberServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Jobber_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(JobberServer).Stream(m, &jobberStreamServer{stream})
}

type Jobber_StreamServer interface {
	Send(*StreamResponse) error
	grpc.ServerStream
}

type jobberStreamServer struct {
	grpc.ServerStream
}

func (x *jobberStreamServer) Send(m *StreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

// Jobber_ServiceDesc is the grpc.ServiceDesc for Jobber service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Jobber_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Jobber",
	HandlerType: (*JobberServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Start",
			Handler:    _Jobber_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Jobber_Stop_Handler,
		},
		{
			MethodName: "Status",
			Handler:    _Jobber_Status_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _Jobber_Stream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "jobber.proto",
}