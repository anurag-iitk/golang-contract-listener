package blockchain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

// DepositEther deposits Ether into the contract
func DepositEther(client *ethclient.Client, contractAddress common.Address, auth *bind.TransactOpts, amount *big.Float) (*types.Transaction, error) {
	contractABI, err := LoadABI()
	if err != nil {
		return nil, err
	}

	// Convert Ether amount to Wei
	weiAmount := new(big.Int)
	amount.Mul(amount, big.NewFloat(params.Ether)).Int(weiAmount)

	// Set the transaction value in Wei
	auth.Value = weiAmount

	tx, err := bind.NewBoundContract(contractAddress, contractABI, client, client, client).
		Transact(auth, "depositEther")
	if err != nil {
		return nil, fmt.Errorf("failed to deposit Ether: %v", err)
	}
	return tx, nil
}
