package blockchain

import (
	"fmt"
	"math/big"

	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ApproveProposal approves a proposal on the blockchain
func ApproveProposal(client *ethclient.Client, contractAddress common.Address, proposalID *big.Int, auth *bind.TransactOpts) (*types.Transaction, error) {
	contractABI, err := LoadABI()
	if err != nil {
		return nil, err
	}

	tx, err := bind.NewBoundContract(contractAddress, contractABI, client, client, client).
		Transact(auth, "approveProposal", proposalID)
	if err != nil {
		return nil, fmt.Errorf("failed to approve proposal: %v", err)
	}
	return tx, nil
}

// GetContractAddress returns the contract address from the environment
func GetContractAddress() common.Address {
	return common.HexToAddress(os.Getenv("DIAMOND_ADDRESS"))
}
