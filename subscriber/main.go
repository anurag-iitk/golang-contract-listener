package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/streadway/amqp"
)

var rabbitConn *amqp.Connection
var rabbitChannel *amqp.Channel

// connectToRabbitMQ establishes a persistent connection to RabbitMQ.
func connectToRabbitMQ() (*amqp.Connection, error) {
	var err error
	if rabbitConn == nil || rabbitConn.IsClosed() {
		// Attempt to connect to RabbitMQ
		rabbitConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			return nil, err
		}
		log.Println("Successfully connected to RabbitMQ")
	}
	return rabbitConn, nil
}

// openRabbitMQChannel ensures the channel is open and persistent.
func openRabbitMQChannel() (*amqp.Channel, error) {
	var err error
	if rabbitChannel == nil {
		rabbitChannel, err = rabbitConn.Channel()
		if err != nil {
			return nil, err
		}
		log.Println("Successfully opened RabbitMQ channel")
	}
	return rabbitChannel, nil
}

// consumeMessages continuously listens for messages from the queue.
func consumeMessages() {
	var err error

	// Ensure connection to RabbitMQ
	_, err = connectToRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// Ensure a channel is open
	_, err = openRabbitMQChannel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	// Declare the queue to consume from
	q, err := rabbitChannel.QueueDeclare(
		"events_queue", // Queue name (should match the publisher's queue)
		true,           // Durable (ensures messages survive restarts)
		false,          // Delete when unused
		false,          // Exclusive
		false,          // No-wait
		nil,            // Arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Consume messages
	msgs, err := rabbitChannel.Consume(
		q.Name, // Queue name
		"",     // Consumer tag
		true,   // Auto-acknowledge messages
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Listening for messages...")

	// Listen for messages
	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)
		}
	}()

	// Block the function from returning
	select {} // Prevents the function from exiting
}

func main() {
	app := fiber.New()

	// Start consuming messages in a goroutine
	go consumeMessages()

	// Define a simple HTTP endpoint for checking the subscriber status
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Subscriber is running")
	})

	// Start the Fiber app
	log.Fatal(app.Listen(":3001"))
}
