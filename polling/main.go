package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Event structures based on the ABI
type InitializedEvent struct {
	Version uint64
}

type ThresholdSetEvent struct {
	Threshold *big.Int
}

// ABI of the smart contract (replace with your actual contract ABI)
const contractABI = `[
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

// Contract address (replace with your actual contract address)
const contractAddress = "0x03b67bE2c5c0CCC16Cb45aa6529111b9fcDaE446"

// Ethereum node RPC URL (replace with your Infura or Geth RPC endpoint)
const rpcURL = "ws://127.0.0.1:8545"

func main() {
	// Connect to Ethereum client
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	// Load contract ABI
	contractAbi, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	// Polling mechanism: checking for new events every 15 seconds
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Keep track of the last block to avoid processing the same block repeatedly
	var lastProcessedBlock uint64

	for range ticker.C {
		// Get the latest block number
		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("Failed to get latest block number: %v", err)
		}

		// Skip if the block hasn't changed since the last poll
		if blockNumber == lastProcessedBlock {
			fmt.Println("No new blocks. Skipping...")
			continue
		}

		// Query logs for the smart contract event
		query := ethereum.FilterQuery{
			Addresses: []common.Address{common.HexToAddress(contractAddress)},
			FromBlock: big.NewInt(int64(lastProcessedBlock + 1)),
			ToBlock:   big.NewInt(int64(blockNumber)),
		}

		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to filter logs: %v", err)
		}

		// Process the logs and decode events
		for _, vLog := range logs {
			switch vLog.Topics[0].Hex() {
			case contractAbi.Events["Initialized"].ID.Hex():
				event := new(InitializedEvent)
				err := contractAbi.UnpackIntoInterface(event, "Initialized", vLog.Data)
				if err != nil {
					log.Fatalf("Failed to unpack Initialized event: %v", err)
				}
				fmt.Printf("Initialized Event detected! Version: %d\n", event.Version)
			case contractAbi.Events["ThresholdSet"].ID.Hex():
				event := new(ThresholdSetEvent)
				err := contractAbi.UnpackIntoInterface(event, "ThresholdSet", vLog.Data)
				if err != nil {
					log.Fatalf("Failed to unpack ThresholdSet event: %v", err)
				}
				fmt.Printf("ThresholdSet Event detected! Threshold: %s\n", event.Threshold.String())
			default:
				fmt.Println("Unknown event detected.")
			}

			// Also print the block and transaction details
			fmt.Printf("Block Number: %d\n", vLog.BlockNumber)
			fmt.Printf("Tx Hash: %s\n", vLog.TxHash.Hex())
		}
		lastProcessedBlock = blockNumber
	}
}
