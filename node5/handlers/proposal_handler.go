package handlers

import (
	"fmt"
	"log"
	"math/big"
	"node1/blockchain"
	"node1/utils"

	"github.com/gofiber/fiber/v2"
)

// ApproveProposalHandler handles the approval of a proposal
func ApproveProposalHandler(c *fiber.Ctx) error {
	// Parse the JSON body
	data := struct {
		ProposalID int64 `json:"proposalId"`
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

	// Approve Proposal
	proposalID := big.NewInt(data.ProposalID)
	fmt.Printf("Approving proposal with ID: %d...\n", proposalID)
	approveProposalTx, err := blockchain.ApproveProposal(client, contractAddress, proposalID, auth)
	if err != nil {
		log.Printf("Failed to approve proposal: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to approve proposal", err)
	}

	// Respond with transaction details
	return c.JSON(fiber.Map{
		"message":       "Proposal approved successfully",
		"transactionID": approveProposalTx.Hash().Hex(),
	})
}
