package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/username/chatapp/common"
	"github.com/username/chatapp/recipient"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := recipient.NewChatServiceClient(conn)

	stream, err := client.BroadcastMessage(context.Background())
	if err != nil {
		log.Fatalf("Error receiving message: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error receiving message: %v", err)
		}

		message := &common.Message{}
		err = proto.Unmarshal(resp.GetData(), message)
		if err != nil {
			log.Fatalf("Error decoding message: %v", err)
		}

		fmt.Printf("[%s]: %s\n", message.GetUsername(), message.GetText())
	}
}
