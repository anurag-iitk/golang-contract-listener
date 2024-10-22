package blockchain

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

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

type ApprovalRequest struct {
	ProposalID int64  `json:"proposalId"`
	PrivateKey string `json:"privateKey"`
}

type DepositRequest struct {
	PrivateKey string `json:"privateKey"`
	Amount     string `json:"amount"`
}

func DepositEther(req DepositRequest) error {
	client, err := ethclient.Dial(os.Getenv("RPC_URL"))
	if err != nil {
		return err
	}
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(req.PrivateKey, "0x"))
	if err != nil {
		return err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337)) // Adjust ChainID if needed
	if err != nil {
		return err
	}
	etherAmount := new(big.Float)
	if _, ok := etherAmount.SetString(req.Amount); !ok {
		return fmt.Errorf("invalid amount format")
	}
	weiAmount := new(big.Float).Mul(etherAmount, big.NewFloat(1e18))
	wei := new(big.Int)
	weiAmount.Int(wei)
	auth.Value = wei
	contractAddress := common.HexToAddress(os.Getenv("DIAMOND_ADDRESS"))
	contractAbi, err := abi.JSON(strings.NewReader(proposalABI))
	if err != nil {
		return err
	}
	contract := bind.NewBoundContract(contractAddress, contractAbi, client, client, client)
	tx, err := contract.Transact(auth, "depositEther")
	if err != nil {
		return err
	}
	log.Printf("Ether deposited with tx: %s", tx.Hash().Hex())
	return nil
}

func ApproveOnChain(client *ethclient.Client, auth *bind.TransactOpts, contractAddress common.Address, proposalID *big.Int) (*types.Transaction, error) {
	contractAbi, err := abi.JSON(strings.NewReader(proposalABI))
	if err != nil {
		return nil, err
	}
	contract := bind.NewBoundContract(contractAddress, contractAbi, client, client, client)
	tx, err := contract.Transact(auth, "approveProposal", proposalID)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func ApproveProposal(req ApprovalRequest) error {
	client, err := ethclient.Dial(os.Getenv("RPC_URL"))
	if err != nil {
		return err
	}
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(req.PrivateKey, "0x"))
	if err != nil {
		return err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337)) // Adjust ChainID
	if err != nil {
		return err
	}
	proposalID := big.NewInt(req.ProposalID)
	contractAddress := common.HexToAddress(os.Getenv("DIAMOND_ADDRESS"))
	tx, err := ApproveOnChain(client, auth, contractAddress, proposalID)
	if err != nil {
		return err
	}
	log.Printf("Proposal approved with tx: %s", tx.Hash().Hex())
	return nil
}
