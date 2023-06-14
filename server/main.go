package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	chat "github.com/AlexanderLukashuk/chatapp/server/proto"
)

type chatServer struct {
	chat.UnimplementedChatServiceServer
}

func (s *chatServer) Broadcast(stream chat.ChatService_BroadcastServer) error {
	log.Println("New client connected for broadcasting")

	s.sendPreviousMessages(stream)

	for {
		message, err := stream.Recv()
		if err != nil {
			return err
		}

		// message.Timestamp = time.Now().Unix()

		if err := s.publishMessage(message); err != nil {
			log.Printf("Failed to publish message: %v", err)
		}

		s.broadcastMessage(message)
	}
}

func (s *chatServer) SendMessage(ctx context.Context, message *chat.Message) (*chat.Message, error) {
	// message.Timestamp = time.Now().Unix()

	if err := s.publishMessage(message); err != nil {
		log.Printf("Failed to publish message: %v", err)
		return nil, err
	}

	return message, nil
}

func (s *chatServer) sendPreviousMessages(stream chat.ChatService_BroadcastServer) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"chat_messages", // Queue name
		false,           // Durable
		false,           // Delete when unused
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // Queue
		"",     // Consumer
		true,   // Auto-ack
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		var message chat.Message
		if err := proto.Unmarshal(msg.Body, &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		if err := stream.Send(&message); err != nil {
			log.Printf("Failed to send previous message: %v", err)
			continue
		}
	}
}

func (s *chatServer) broadcastMessage(message *chat.Message) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"chat_messages", // Queue name
		false,           // Durable
		false,           // Delete when unused
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	messageBytes, err := proto.Marshal(message)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v", err)
	}

	err = ch.Publish(
		"",     // Exchange
		q.Name, // Routing key
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        messageBytes,
		})
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
}

func (s *chatServer) publishMessage(message *chat.Message) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"chat_messages", // Queue name
		false,           // Durable
		false,           // Delete when unused
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	err = ch.Publish(
		"",     // Exchange
		q.Name, // Routing key
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        messageBytes,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	chatSrv := chatServer{}
	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, &chatSrv)

	log.Println("Chat server started")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
