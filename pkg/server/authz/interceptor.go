package authz

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func (r *RBAC) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no peer found")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, fmt.Errorf("unexpected peer transport credentials")
	}

	// Check the client certificates
	for _, cert := range tlsAuth.State.PeerCertificates {
		// Perform any additional checks on the client certificate here
		for _, addr := range cert.EmailAddresses {
			if r.Authorize(addr, info.FullMethod) {
				// Continue handling the request
				return handler(ctx, req)
			}
		}
	}

	return nil, fmt.Errorf("not authorized")
}

func (r *RBAC) StreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	p, ok := peer.FromContext(ss.Context())
	if !ok {
		return fmt.Errorf("no peer found")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return fmt.Errorf("unexpected peer transport credentials")
	}

	// Check the client certificates
	for _, cert := range tlsAuth.State.PeerCertificates {
		// Perform any additional checks on the client certificate here
		for _, addr := range cert.EmailAddresses {
			if r.Authorize(addr, info.FullMethod) {
				// Continue handling the request
				return handler(srv, ss)
			}
		}
	}

	return fmt.Errorf("not authorized")
}
