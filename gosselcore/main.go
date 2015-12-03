package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/ayonix/gossel/gosselproto"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 4343, "Port the Server listens on")

func main() {
	flag.Parse()

	c := NewGosselcore()
	c.AddUser("Test", "Test")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGosselcoreServer(grpcServer, &gosselcoreServer{c})
	grpcServer.Serve(lis)

	fmt.Printf("Listening on port %d \n", port)
}
