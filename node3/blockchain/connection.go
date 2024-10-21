package blockchain

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// ConnectToBlockchain connects to the Ethereum blockchain using RPC
func ConnectToBlockchain() (*ethclient.Client, *rpc.Client, error) {
	rpcURL := os.Getenv("RPC_URL")
	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}
	client := ethclient.NewClient(rpcClient)
	return client, rpcClient, nil
}
