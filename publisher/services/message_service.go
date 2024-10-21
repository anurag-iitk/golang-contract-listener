package services

import (
	"encoding/json"
	"fmt"
	"log"
	"publisher/config"

	"github.com/streadway/amqp"
)

func PublishEventToRabbitMQ(eventPayload interface{}) error {
	_, ch, err := config.ConnectRabbitMQ()
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	payloadJSON, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("failed to serialize event payload: %w", err)
	}

	err = ch.Publish(
		"",
		"events_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(payloadJSON),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event to RabbitMQ: %w", err)
	}

	log.Printf("Event successfully published to RabbitMQ: %s", string(payloadJSON))
	return nil
}
