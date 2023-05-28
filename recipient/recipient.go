package main

import (
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/username/chatapp/common"
	"github.com/username/chatapp/server"
	"google.golang.org/grpc"
)

type chatServer struct{}

func (s *chatServer) BroadcastMessage(stream server.ChatService_BroadcastMessageServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error receiving message: %v", err)
		}

		message := &common.Message{}
		err = proto.Unmarshal(req.GetData(), message)
		if err != nil {
			log.Fatalf("Error decoding message: %v", err)
		}

		fmt.Printf("[%s]: %s\n", message.GetUsername(), message.GetText())
	}
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	server := grpc.NewServer()

	server.RegisterChatServiceServer(server, &chatServer{})

	log.Println("Server started, listening on port 50051")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
