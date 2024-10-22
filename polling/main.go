package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
)

type EventResponse struct {
	BlockNumber uint64 `json:"block_number"`
	TxHash      string `json:"tx_hash"`
	EventName   string `json:"event_name"`
	Details     string `json:"details"`
}

var eventLog []EventResponse // Holds a log of detected events

// ABI definitions
const InitializerABI = `[
    {
      "inputs": [],
      "name": "InvalidInitialization",
      "type": "error"
    },
    {
      "inputs": [],
      "name": "NotInitializing",
      "type": "error"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "internalType": "uint64",
          "name": "version",
          "type": "uint64"
        }
      ],
      "name": "Initialized",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "internalType": "uint256",
          "name": "threshold",
          "type": "uint256"
        }
      ],
      "name": "ThresholdSet",
      "type": "event"
    },
    {
      "inputs": [
        {
          "internalType": "uint256",
          "name": "_threshold",
          "type": "uint256"
        }
      ],
      "name": "initialize",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ]`

const approvalABI = `[
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "internalType": "address",
          "name": "approver",
          "type": "address"
        }
      ],
      "name": "ApproverAdded",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "internalType": "address",
          "name": "approver",
          "type": "address"
        }
      ],
      "name": "ApproverDeleted",
      "type": "event"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_approver",
          "type": "address"
        }
      ],
      "name": "addApprover",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_approver",
          "type": "address"
        }
      ],
      "name": "deleteApprover",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address",
          "name": "_approver",
          "type": "address"
        }
      ],
      "name": "getApprover",
      "outputs": [
        {
          "internalType": "bool",
          "name": "",
          "type": "bool"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    }
  ]`

const proposalABI = `[
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "internalType": "address",
          "name": "sender",
          "type": "address"
        },
        {
          "indexed": false,
          "internalType": "uint256",
          "name": "amount",
          "type": "uint256"
        }
      ],
      "name": "DepositedEther",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "internalType": "address",
          "name": "recipient",
          "type": "address"
        },
        {
          "indexed": false,
          "internalType": "uint256",
          "name": "amount",
          "type": "uint256"
        }
      ],
      "name": "ProposalAdded",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "internalType": "uint256",
          "name": "proposalId",
          "type": "uint256"
        }
      ],
      "name": "ProposalApproved",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "internalType": "uint256",
          "name": "_proposalId",
          "type": "uint256"
        }
      ],
      "name": "ProposalExecuted",
      "type": "event"
    },
    {
      "inputs": [
        {
          "internalType": "uint256",
          "name": "_proposalId",
          "type": "uint256"
        }
      ],
      "name": "approveProposal",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "address payable",
          "name": "_recipient",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "_amount",
          "type": "uint256"
        }
      ],
      "name": "createProposal",
      "outputs": [],
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [],
      "name": "depositEther",
      "outputs": [],
      "stateMutability": "payable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "internalType": "uint256",
          "name": "_proposalId",
          "type": "uint256"
        }
      ],
      "name": "getProposal",
      "outputs": [
        {
          "internalType": "uint256",
          "name": "proposalId",
          "type": "uint256"
        },
        {
          "internalType": "address",
          "name": "recipient",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "amount",
          "type": "uint256"
        },
        {
          "internalType": "uint256",
          "name": "approvals",
          "type": "uint256"
        },
        {
          "internalType": "bool",
          "name": "executed",
          "type": "bool"
        }
      ],
      "stateMutability": "view",
      "type": "function"
    }
  ]`

// Track multiple contracts with their ABIs and addresses
var contracts = map[string]string{
	"Initializer": InitializerABI,
	"Approval":    approvalABI,
	"Proposal":    proposalABI,
}

func pollEvents(client *ethclient.Client, contractAbi abi.ABI, contractAddress string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastProcessedBlock uint64

	for range ticker.C {
		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Printf("Failed to get latest block number: %v", err)
			continue
		}

		if blockNumber == lastProcessedBlock {
			continue
		}

		query := ethereum.FilterQuery{
			Addresses: []common.Address{common.HexToAddress(contractAddress)},
			FromBlock: big.NewInt(int64(lastProcessedBlock + 1)),
			ToBlock:   big.NewInt(int64(blockNumber)),
		}

		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Printf("Failed to filter logs: %v", err)
			continue
		}

		for _, vLog := range logs {
			eventResponse := EventResponse{
				BlockNumber: vLog.BlockNumber,
				TxHash:      vLog.TxHash.Hex(),
			}

			// Process each event for the ABI
			for name, event := range contractAbi.Events {
				if vLog.Topics[0].Hex() == event.ID.Hex() {
					eventResponse.EventName = name
					details, err := processEvent(contractAbi, event, vLog)
					if err != nil {
						log.Printf("Failed to process event %s: %v", name, err)
						continue
					}
					eventResponse.Details = details
					break
				}
			}

			if eventResponse.EventName != "" {
				// Log and store the event only if EventName is populated
				log.Printf("Event detected: %s, Block: %d, Tx: %s\n", eventResponse.EventName, eventResponse.BlockNumber, eventResponse.TxHash)
				eventLog = append(eventLog, eventResponse)
			}
		}

		lastProcessedBlock = blockNumber
	}
}

// processEvent dynamically processes the events based on ABI and log data
func processEvent(contractAbi abi.ABI, event abi.Event, vLog types.Log) (string, error) {
	// Create a map to unpack event data into
	eventData := map[string]interface{}{}
	err := contractAbi.UnpackIntoMap(eventData, event.Name, vLog.Data)
	if err != nil {
		return "", err
	}

	// Generate a human-readable output for the event
	var details []string
	for name, value := range eventData {
		details = append(details, fmt.Sprintf("%s: %v", name, value))
	}
	return strings.Join(details, ", "), nil
}

func main() {
	rpcURL := os.Getenv("RPC_URL")
	contractAddress := os.Getenv("CONTRACT_ADDRESS")

	// Connect to Ethereum client
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	// Start polling for each contract ABI
	for contractName, contractABI := range contracts {
		contractAbi, err := abi.JSON(strings.NewReader(contractABI))
		if err != nil {
			log.Fatalf("Failed to parse %s contract ABI: %v", contractName, err)
		}

		// Start polling for events in a separate goroutine for each contract
		go pollEvents(client, contractAbi, contractAddress)
	}

	// Start Fiber server
	app := fiber.New()

	// Route to check detected events
	app.Get("/events", func(c *fiber.Ctx) error {
		return c.JSON(eventLog)
	})

	// Start the Fiber server on port 4003
	log.Fatal(app.Listen(":4003"))
}
