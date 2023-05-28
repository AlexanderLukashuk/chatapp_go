package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/username/chatapp/common"
	"github.com/username/chatapp/server"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := server.NewChatServiceClient(conn)

	fmt.Print("Enter your name: ")
	senderName := readLine()

	stream, err := client.BroadcastMessage(context.Background())
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

	for {
		fmt.Print("Enter message: ")
		text := readLine()

		message := &common.Message{
			Username: senderName,
			Text:     text,
		}

		data, err := proto.Marshal(message)
		if err != nil {
			log.Fatalf("Error encoding message: %v", err)
		}

		err = stream.Send(&server.MessageRequest{Data: data})
		if err != nil {
			log.Fatalf("Error sending message: %v", err)
		}
	}
}

func readLine() string {
	var input string
	fmt.Scanln(&input)
	return input
}
