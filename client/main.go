package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	chat "github.com/AlexanderLukashuk/chatapp/server/proto"
)

func receiveMessages(stream chat.ChatService_BroadcastClient, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		message, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			return
		}

		// timestamp := message.GetTimestamp()
		username := message.GetUsername()
		content := message.GetContent()
		// t := timestampToTime(timestamp)
		// log.Printf("%s: %s: %s", username, content)
		log.Printf("%s: %s", username, content)
	}
}

func timestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func doSendMessage(c chat.ChatServiceClient, msg *chat.Message) {
	ctx := context.Background()

	request := &chat.Request{Message: msg}
	response, err := c.SendMessage(ctx, request.Message)

	if err != nil {
		log.Fatalf("Error while calling send RPC %v", err)
	}

	// timestamp := response.GetTimestamp()
	username := response.GetUsername()
	content := response.GetContent()
	// t := timestampToTime(timestamp)
	// log.Printf("%s: %s: %s", t.Format("02/01/2006 15:04"), username, content)
	log.Printf("%s: %s", username, content)
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
	username = strings.TrimSpace(username)

	fmt.Println("Type your message and press Enter to send. Type 'quit' to exit.")

	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "quit" {
			break
		}

		doSendMessage(client, &chat.Message{
			Username: username,
			Content:  message,
		})

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
