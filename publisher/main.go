// main.go
package main

import (
	"log"
	"publisher/blockchain"
	"publisher/utils"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize logger
	logger := utils.NewLogger()

	// Create a new Fiber app
	app := fiber.New()

	// Start listening to blockchain events in a separate goroutine
	go blockchain.ListenToContractEvents()

	// Start the Fiber app
	logger.Info("Starting server on port 3000...")
	log.Fatal(app.Listen(":3000"))
}
