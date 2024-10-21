package handlers

import (
	"fmt"
	"log"
	"math/big"
	"node1/blockchain"
	"node1/utils"

	"github.com/gofiber/fiber/v2"
)

// DepositEtherHandler handles depositing Ether to the contract
func DepositEtherHandler(c *fiber.Ctx) error {
	// Parse the JSON body for the deposit amount
	data := struct {
		Amount string `json:"amount"`
	}{}

	if err := c.BodyParser(&data); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request", err)
	}

	// Connect to blockchain
	client, _, err := blockchain.ConnectToBlockchain()
	if err != nil {
		log.Printf("Blockchain connection error: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Blockchain connection error", err)
	}
	defer client.Close()

	// Prepare auth
	auth, err := blockchain.GetAuth(client)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Authentication error", err)
	}

	// Get contract address
	contractAddress := blockchain.GetContractAddress()

	// Parse deposit amount and call the deposit function
	depositAmount, ok := new(big.Float).SetString(data.Amount)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid deposit amount", nil)
	}
	fmt.Printf("Depositing %s Ether...\n", data.Amount)
	depositTx, err := blockchain.DepositEther(client, contractAddress, auth, depositAmount)
	if err != nil {
		log.Printf("Failed to deposit Ether: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to deposit Ether", err)
	}

	// Respond with transaction details
	return c.JSON(fiber.Map{
		"message":       "Deposit successful",
		"transactionID": depositTx.Hash().Hex(),
	})
}
