package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"google.golang.org/grpc"

	chat "github.com/AlexanderLukashuk/chatapp/server/proto" // Import the generated protobuf code
)

func receiveMessages(stream chat.ChatService_BroadcastClient, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		message, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			return
		}

		log.Printf("[%s]: %s", message.Username, message.Content)
	}
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := chat.NewChatServiceClient(conn)

	stream, err := client.Broadcast(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go receiveMessages(stream, &wg)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your username: ")
	username, _ := reader.ReadString('\n')

	fmt.Println("Type your message and press Enter to send. Type 'quit' to exit.")

	for {
		message, _ := reader.ReadString('\n')
		message = message[:len(message)-1] // Remove newline character

		if message == "quit" {
			break
		}

		// Send the message to the server
		if err := stream.Send(&chat.Message{
			Username: username,
			Content:  message,
		}); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}

	stream.CloseSend()
	wg.Wait()
}
