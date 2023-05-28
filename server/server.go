package main

import (
	"fmt"
	"log"
	"net"

	"github.com/username/chatapp/common"
	"github.com/username/chatapp/rabbitmq"
	"github.com/username/chatapp/server"
	"go.starlark.net/lib/proto"
	"google.golang.org/grpc"
)

type chatServer struct {
	rabbitMQ *rabbitmq.RabbitMQ
}

func (s *chatServer) BroadcastMessage(stream server.ChatService_BroadcastMessageServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("error receiving message: %v", err)
		}

		message := &common.Message{}
		err = proto.Unmarshal(req.GetData(), message)
		if err != nil {
			return fmt.Errorf("error decoding message: %v", err)
		}

		err = s.rabbitMQ.PublishMessage(req.GetData())
		if err != nil {
			return fmt.Errorf("error publishing message: %v", err)
		}
	}
}

func main() {
	rabbitMQ, err := rabbitmq.NewRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	server := grpc.NewServer()
	chatServiceServer := &chatServer{
		rabbitMQ: rabbitMQ,
	}

	server.RegisterChatServiceServer(server, chatServiceServer)

	log.Println("Server started, listening on port 50051")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
