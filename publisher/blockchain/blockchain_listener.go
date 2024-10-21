package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"publisher/services"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// Event ABI
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

type EventPayload struct {
	EventName        string                 `json:"eventName"`
	Data             map[string]interface{} `json:"data"`
	BlockNumber      uint64                 `json:"blockNumber"`
	TransactionHash  string                 `json:"transactionHash"`
	BlockHash        string                 `json:"blockHash"`
	TransactionIndex uint                   `json:"transactionIndex"`
	LogIndex         uint                   `json:"logIndex"`
	Address          string                 `json:"address"`
}

var combinedABIs = []string{InitializerABI, approvalABI, proposalABI}

func getCombinedABI() (abi.ABI, error) {
	parsedABI := abi.ABI{
		Methods: make(map[string]abi.Method),
		Events:  make(map[string]abi.Event),
	}

	for _, abiString := range combinedABIs {
		tempABI, err := abi.JSON(strings.NewReader(abiString))
		if err != nil {
			return parsedABI, fmt.Errorf("failed to parse ABI: %v", err)
		}
		parsedABI = appendABI(parsedABI, tempABI)
	}

	return parsedABI, nil
}

func appendABI(dest abi.ABI, src abi.ABI) abi.ABI {
	for name, event := range src.Events {
		dest.Events[name] = event
	}
	return dest
}

func createEventSignatureMap(contractABI abi.ABI) map[string]abi.Event {
	eventMap := make(map[string]abi.Event)
	for _, event := range contractABI.Events {
		eventMap[event.ID.Hex()] = event
	}
	return eventMap
}

func ListenToContractEvents() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	contractAddress := os.Getenv("CONTRACT_ADDRESS")
	if contractAddress == "" {
		log.Fatalf("CONTRACT_ADDRESS environment variable not set")
	}

	rpcUrl := os.Getenv("RPC_URL")
	if contractAddress == "" {
		log.Fatalf("CONTRACT_ADDRESS environment variable not set")
	}

	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	log.Println("Connected to Ethereum node via WebSocket.")

	combinedABI, err := getCombinedABI()
	if err != nil {
		log.Fatalf("Failed to get combined ABI: %v", err)
	}
	eventSignatureMap := createEventSignatureMap(combinedABI)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to event logs: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Successfully subscribed to contract events.")
	for {
		select {
		case err := <-sub.Err():
			log.Printf("Error while listening to event logs: %v. Reconnecting...", err)
			time.Sleep(5 * time.Second)
			return
		case vLog := <-logs:
			event, ok := eventSignatureMap[vLog.Topics[0].Hex()]
			if !ok {
				log.Printf("Unknown event signature: %s", vLog.Topics[0].Hex())
				continue
			}

			payload, err := createEventPayload(&event, combinedABI, vLog)
			if err != nil {
				log.Printf("Failed to parse log data: %v", err)
				continue
			}
			err = services.PublishEventToRabbitMQ(payload)
			if err != nil {
				log.Printf("Failed to publish event to RabbitMQ: %v", err)
			}
		}
	}
}

func createEventPayload(event *abi.Event, contractAbi abi.ABI, vLog types.Log) (EventPayload, error) {
	dataMap := make(map[string]interface{})
	err := contractAbi.UnpackIntoMap(dataMap, event.Name, vLog.Data)
	if err != nil {
		return EventPayload{}, fmt.Errorf("failed to unpack data: %v", err)
	}

	for i, input := range event.Inputs {
		if input.Indexed {
			dataMap[input.Name] = parseTopicValue(input.Type, vLog.Topics[i+1])
		}
	}

	payload := EventPayload{
		EventName:        event.Name,
		Data:             dataMap,
		BlockNumber:      vLog.BlockNumber,
		TransactionHash:  vLog.TxHash.Hex(),
		BlockHash:        vLog.BlockHash.Hex(),
		TransactionIndex: vLog.TxIndex,
		LogIndex:         vLog.Index,
		Address:          vLog.Address.Hex(),
	}

	return payload, nil
}

func parseTopicValue(argType abi.Type, topic common.Hash) interface{} {
	switch argType.T {
	case abi.AddressTy:
		return common.HexToAddress(topic.Hex()).Hex()
	case abi.UintTy, abi.IntTy:
		return new(big.Int).SetBytes(topic[:]).String()
	default:
		return topic.Hex()
	}
}
