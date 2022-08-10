package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	DataDir          = "data"
	BlockNum         = 15129365
	TransactionsRoot = "0x6be9be79ba3847cc77f5ec61747bf2fd888474631bc0dded9e2c455e17994c36"
	ReceiptsRoot     = " 0x0595fa2fea554388f3ac06cb67ddc1e644fea876cb0e9d14fac30d266602afe9"
)

func TransactionsFromJSON() []*types.Transaction {
	transactionsFile := DataDir + "/transactions-" + strconv.Itoa(BlockNum) + ".json"
	jsonFile, err := os.Open(transactionsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	var txs []*types.Transaction
	json.Unmarshal(byteValue, &txs)
	return txs
}

func ReceiptsFromJSON() []*types.Receipt {
	receiptsFile := DataDir + "/receipts-" + strconv.Itoa(BlockNum) + ".json"
	jsonFile, err := os.Open(receiptsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}
	var receipts []*types.Receipt
	json.Unmarshal(byteValue, &receipts)
	return receipts
}

func TestTransactionsRoot() {
	txs := TransactionsFromJSON()

	hasher := trie.NewStackTrie(nil)
	treeHash := types.DeriveSha(types.Transactions(txs), hasher)
	transactionRoot := "0x6be9be79ba3847cc77f5ec61747bf2fd888474631bc0dded9e2c455e17994c36"

	fmt.Println("expected transactions root: ", transactionRoot)
	fmt.Println("transactions root from tree: ", treeHash.String())
	if transactionRoot == treeHash.String() {
		fmt.Println("roots match!")
	}
}

func TestReceiptsRoot() {
	receipts := ReceiptsFromJSON()

	hasher := trie.NewStackTrie(nil)
	treeHash := types.DeriveSha(types.Receipts(receipts), hasher)

	receiptsRoot := "0x0595fa2fea554388f3ac06cb67ddc1e644fea876cb0e9d14fac30d266602afe9"
	fmt.Println("expected receipts root: ", receiptsRoot)
	fmt.Println("receipts root from tree: ", treeHash.String())
	if receiptsRoot == treeHash.String() {
		fmt.Println("roots match!")
	}
}

func main() {
	TestTransactionsRoot()
	TestReceiptsRoot()
}
