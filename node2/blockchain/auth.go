package blockchain

import (
	"fmt"
	"os"
	"strings"

	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetAuth prepares authentication using a private key from the environment
func GetAuth(client *ethclient.Client) (*bind.TransactOpts, error) {
	privateKey := os.Getenv("APPROVER_PRIVATE_KEY")
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	chainID := big.NewInt(31337)
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorized transactor: %v", err)
	}
	return auth, nil
}
