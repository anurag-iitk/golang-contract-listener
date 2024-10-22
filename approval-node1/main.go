package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

type RabbitMQHandler struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// NewRabbitMQHandler initializes a single RabbitMQ handler.
func NewRabbitMQHandler() (*RabbitMQHandler, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQHandler{
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (r *RabbitMQHandler) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Connection != nil {
		r.Connection.Close()
	}
}

// SendRequest is a generic function to send a request to the specified queue and wait for a response.
func (r *RabbitMQHandler) SendRequest(requestBody map[string]interface{}, routingKey string) (string, <-chan amqp.Delivery, error) {
	callbackQueue, err := r.Channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return "", nil, err
	}

	msgs, err := r.Channel.Consume(
		callbackQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return "", nil, err
	}

	correlationID := uuid.New().String()
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", nil, err
	}

	err = r.Channel.Publish(
		"",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          bodyBytes,
			CorrelationId: correlationID,
			ReplyTo:       callbackQueue.Name,
		},
	)
	if err != nil {
		return "", nil, err
	}

	return correlationID, msgs, nil
}

// HandleRequest is a generic function to handle both "approve" and "deposit" requests.
func HandleRequest(c *fiber.Ctx, rabbitMQ *RabbitMQHandler, routingKey string, requestBody map[string]interface{}) error {
	correlationID, msgs, err := rabbitMQ.SendRequest(requestBody, routingKey)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	for msg := range msgs {
		if msg.CorrelationId == correlationID {
			log.Printf("Received response: %s", msg.Body)
			var response map[string]string
			err = json.Unmarshal(msg.Body, &response)
			if err != nil {
				log.Printf("Failed to unmarshal response: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Failed to parse response",
					"details": string(msg.Body),
				})
			}
			return c.JSON(fiber.Map{
				"response": response,
			})
		}
	}

	return c.JSON(fiber.Map{"message": "No response received"})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	rabbitMQ, err := NewRabbitMQHandler()
	if err != nil {
		log.Fatalf("RabbitMQ initialization error: %v", err)
	}
	defer rabbitMQ.Close()

	app := fiber.New()

	// Approve route
	app.Post("/approve", func(c *fiber.Ctx) error {
		privateKey := os.Getenv("APPROVER_PRIVATE_KEY")
		proposalID := int64(1)
		requestBody := map[string]interface{}{
			"proposalId": proposalID,
			"privateKey": privateKey,
		}
		return HandleRequest(c, rabbitMQ, "approval_queue", requestBody)
	})

	// Deposit route
	app.Post("/deposit", func(c *fiber.Ctx) error {
		type DepositRequest struct {
			Amount string `json:"amount"`
		}
		req := new(DepositRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		privateKey := os.Getenv("APPROVER_PRIVATE_KEY")
		if privateKey == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Private key not set in environment variables"})
		}

		requestBody := map[string]interface{}{
			"privateKey": privateKey,
			"amount":     req.Amount,
		}
		return HandleRequest(c, rabbitMQ, "deposit_queue", requestBody)
	})

	log.Fatal(app.Listen(":4001"))
}
