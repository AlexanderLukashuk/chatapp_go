package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/username/chatapp/rabbitmq"
	"github.com/username/chatapp/server"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		runServer()
	}()

	go func() {
		defer wg.Done()
		runClient()
	}()

	// Wait for termination signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh

	log.Println("Terminating the chat application...")

	// Cleanup and shutdown
	wg.Wait()
}

func runServer() {
	server := server.NewChatServer()
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()
}

func runClient() {
	client := rabbitmq.NewRabbitMQClient()
	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer client.Stop()

	fmt.Print("Enter your name: ")
	senderName := readLine()

	for {
		fmt.Print("Enter message: ")
		text := readLine()

		message := &rabbitmq.Message{
			Username: senderName,
			Text:     text,
		}

		if err := client.PublishMessage(message); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func readLine() string {
	var input string
	fmt.Scanln(&input)
	return input
}
