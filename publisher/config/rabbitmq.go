package config

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

func ConnectRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	var err error
	if conn == nil {
		conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		}
		log.Println("Successfully connected to RabbitMQ")
	} else if conn.IsClosed() {
		return nil, nil, fmt.Errorf("connection to RabbitMQ is closed")
	}

	if channel == nil {
		channel, err = conn.Channel()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
		}
		log.Println("Successfully opened a channel to RabbitMQ")
	}

	_, err = channel.QueueDeclare(
		"events_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return conn, channel, nil
}

func CloseRabbitMQ() {
	if channel != nil {
		err := channel.Close()
		if err != nil {
			log.Printf("Failed to close channel: %v", err)
		}
	}
	if conn != nil {
		err := conn.Close()
		if err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}
}
