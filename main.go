package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log/slog"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		slog.Error("failed to connect to Ethereum client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Import Private Key
	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80") // from Anvil
	if err != nil {
		slog.Error("failed to get private Key", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Get public key from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		slog.Error("error casting public key to ECDSA")
		os.Exit(1)
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		slog.Error("failed to get pending nonce", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Setup TX values
	value := big.NewInt(5000000000000000000) // in wei (5 eth)
	gasLimit := uint64(21000)                // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		slog.Error("failed to get suggested gas price", slog.String("error", err.Error()))
		os.Exit(1)
	}

	toAddress := common.HexToAddress("2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6") // from Anvil
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     []byte{},
	})
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
	if err != nil {
		slog.Error("failed to get signed transaction", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := client.SendTransaction(context.Background(), signedTx); err != nil {
		slog.Error("failed to send transaction", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("Sent %s wei to %s: %s\n", value.String(), toAddress.Hex(), signedTx.Hash().Hex())

	currentPendingFromWalletBalance, err := client.PendingBalanceAt(context.Background(), fromAddress)
	if err != nil {
		slog.Error("failed to get pending from wallet balance", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("Current pending from balance: %s\n", currentPendingFromWalletBalance.String())

	currentPendingToWalletBalance, err := client.PendingBalanceAt(context.Background(), toAddress)
	if err != nil {
		slog.Error("failed to get pending to wallet balance", slog.String("error", err.Error()))
		os.Exit(1)
	}

	fmt.Printf("Current pending to wallet balance: %s\n", currentPendingToWalletBalance.String())
}
