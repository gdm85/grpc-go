package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

// grpcHandlerFunc returns an http.Handler that delegates to
// grpcServer on incoming gRPC connections or otherHandler
// otherwise. Copied from cockroachdb.
func grpcHandlerFunc(rpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("content-type header: %v, method: %v", r.Header.Get("Content-Type"), r.Method)
		if r.ProtoMajor == 2 && (strings.Contains(r.Header.Get("Content-Type"), "application/grpc") || r.Method == "PRI") {
			log.Printf("handling gRPC request")
			rpcServer.ServeHTTP(w, r)
			return
		}

		log.Printf("handling regular HTTP1.x/2 request")
		otherHandler.ServeHTTP(w, r)
	})
}

const port = 50051

func main() {
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: grpcHandlerFunc(s, http.DefaultServeMux),
	}

	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = srv.Serve(lis)
	if err != nil {
		log.Fatalf("serve error: %v", err)
	}
}
