package main

import (
	blockchain "blockchain/services"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

func handleRequest(ch *amqp.Channel, d amqp.Delivery, requestType string) error {
	var err error
	switch requestType {
	case "approval":
		var approvalReq blockchain.ApprovalRequest
		if err = json.Unmarshal(d.Body, &approvalReq); err == nil && approvalReq.ProposalID != 0 {
			log.Printf("Received approval request: %+v", approvalReq)
			err = blockchain.ApproveProposal(approvalReq)
		} else {
			log.Printf("Invalid approval request format: %v", err)
		}
	case "deposit":
		var depositReq blockchain.DepositRequest
		if err = json.Unmarshal(d.Body, &depositReq); err == nil && depositReq.Amount != "" {
			log.Printf("Received deposit request: %+v", depositReq)
			err = blockchain.DepositEther(depositReq)
		} else {
			log.Printf("Invalid deposit request format: %v", err)
		}
	default:
		log.Printf("Unknown request type: %s", requestType)
		err = json.Unmarshal(d.Body, nil)
	}

	if err != nil {
		log.Printf("Failed to process %s request: %v", requestType, err)
	}
	return sendResponse(ch, d, err)
}

func sendResponse(ch *amqp.Channel, d amqp.Delivery, err error) error {
	response := map[string]string{}
	if err != nil {
		response["error"] = err.Error()
	} else {
		response["message"] = "Success"
	}
	responseBody, _ := json.Marshal(response)
	err = ch.Publish(
		"",
		d.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          responseBody,
		},
	)
	if err != nil {
		log.Printf("Failed to publish response: %v", err)
		return err
	}
	log.Printf("Response sent for correlation ID: %s", d.CorrelationId)
	return nil
}

func consumeMessages(ch *amqp.Channel, queueName string, requestType string) {
	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer for %s queue: %v", requestType, err)
	}

	go func() {
		for d := range msgs {
			err := handleRequest(ch, d, requestType)
			if err != nil {
				log.Printf("Failed to handle %s request: %v", requestType, err)
			}
		}
	}()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer ch.Close()

	queues := map[string]string{
		"approval_queue": "approval",
		"deposit_queue":  "deposit",
	}

	for queueName, requestType := range queues {
		_, err := ch.QueueDeclare(
			queueName,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Fatalf("Failed to declare %s queue: %v", requestType, err)
		}

		consumeMessages(ch, queueName, requestType)
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	select {}
}
