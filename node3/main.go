package main

import (
	"log"
	"node1/config"
	"node1/handlers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Create a Fiber app
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Node-3 is working!")
	})

	// Define API endpoints
	app.Post("/approve-proposal", handlers.ApproveProposalHandler)
	app.Post("/deposit-ether", handlers.DepositEtherHandler)

	// Start the Fiber app on port 3004
	log.Fatal(app.Listen(":3004"))
}
