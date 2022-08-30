package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"transactions-and-receipts/simpletrie"
)

const (
	DataDir = "data"

	PreLondonBlockNum = 12964000
	PreLondonTxnsRoot = "0x3259ac1bf20f0e2a02362361bae5489170d25255c18fb07dfe9282c86cecc568"

	PostLondonBlockNum      = 15415840
	PostLondonBlockTxnsRoot = "0xda75f10e5c3ca8adc0c0969da0020377f49f088c436751b735fa8dd4a059ace2"
)

type TrieUpdater interface {
	Update([]byte, []byte)
	Reset()
}

// TrieHasher is the tool used to calculate the hash of derivable list.
type TrieHasher interface {
	Reset()
	Update([]byte, []byte)
	Hash() common.Hash
}

// OldDerivableList is the interface which can derive the hash.
type OldDerivableList interface {
	Len() int
	GetRlp(i int) []byte
}

// DerivableList is the input to DeriveSha.
// New DerivableList type
// It is implemented by the 'Transactions' and 'Receipts' types.
// This is internal, do not use these methods.
type DerivableList interface {
	Len() int
	EncodeIndex(int, *bytes.Buffer)
}

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}

func TransactionsFromJSON(blockNum int) []*types.Transaction {
	transactionsFile := DataDir + "/block-" + strconv.Itoa(blockNum) + "-transactions.json"
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

// encodeBufferPool holds temporary encoder buffers for DeriveSha and TX encoding.
var encodeBufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func encodeForDerive(list DerivableList, i int, buf *bytes.Buffer) []byte {
	buf.Reset()
	list.EncodeIndex(i, buf)
	// It's really unfortunate that we need to do perform this copy.
	// StackTrie holds onto the values until Hash is called, so the values
	// written to it must not alias.
	return common.CopyBytes(buf.Bytes())
}

func InsertTrieIndexOrder(list DerivableList, hasher TrieUpdater) {
	keybuf := new(bytes.Buffer)
	valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
	defer encodeBufferPool.Put(valueBuf)

	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		value := encodeForDerive(list, i, valueBuf)

		hasher.Update(keybuf.Bytes(), value)
	}
}

// OldDeriveSha is the old implementation of DeriveSha.
func OldDeriveSha(list DerivableList, hasher TrieHasher) common.Hash {
	hasher.Reset()

	InsertTrieIndexOrder(list, hasher)
	return hasher.Hash()
}

func InsertTrieByteOrder(list DerivableList, hasher TrieUpdater) {
	hasher.Reset()
	valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
	defer encodeBufferPool.Put(valueBuf)

	// StackTrie requires values to be inserted in increasing hash order, which is not the
	// order that `list` provides hashes in. This insertion sequence ensures that the
	// order is correct.
	var indexBuf []byte
	for i := 1; i < list.Len() && i <= 0x7f; i++ {
		indexBuf = rlp.AppendUint64(indexBuf[:0], uint64(i))
		value := encodeForDerive(list, i, valueBuf)
		hasher.Update(indexBuf, value)
	}
	if list.Len() > 0 {
		indexBuf = rlp.AppendUint64(indexBuf[:0], 0)
		value := encodeForDerive(list, 0, valueBuf)
		hasher.Update(indexBuf, value)
	}
	for i := 0x80; i < list.Len(); i++ {
		indexBuf = rlp.AppendUint64(indexBuf[:0], uint64(i))
		value := encodeForDerive(list, i, valueBuf)
		hasher.Update(indexBuf, value)
	}
}

// DeriveSha creates the tree hashes of transactions and receipts in a block header.
func DeriveSha(list DerivableList, hasher TrieHasher) common.Hash {
	InsertTrieByteOrder(list, hasher)
	return hasher.Hash()
}

func CheckHash(expected string, actual []byte) {
	// expected hash is a hex string, convert to bytes
	expectedBytes, err := hex.DecodeString(expected[2:])
	PanicError(err)

	if !bytes.Equal(expectedBytes, actual) {
		fmt.Println("ROOT HASH DOES NOT MATCH")
	} else {
		fmt.Println("ROOT HASH MATCHES")
	}
	//fmt.Println("expected:", expectedBytes)
	//fmt.Println("actual:", actual)
	fmt.Printf("%-10s: %x\n", "expected", expectedBytes)
	fmt.Printf("%-10s: %x\n", "actual", actual)
	fmt.Println("-------------------------------------------------------")
	fmt.Println("-------------------------------------------------------")
}

// TrieOldShaNewBlock calculates the root hash using the old Trie structure
// along with the old DeriveSha method and the new Block structure.
func TrieOldShaNewBlock(txns []*types.Transaction, expectedRoot string) {
	hasher := new(trie.Trie)
	txnRootHash := OldDeriveSha(types.Transactions(txns), hasher)
	CheckHash(expectedRoot, txnRootHash.Bytes())
}

func TrieNewShaNewBlock(txns []*types.Transaction, expectedRoot string) {
	hasher := new(trie.Trie)
	txnRootHash := DeriveSha(types.Transactions(txns), hasher)
	CheckHash(expectedRoot, txnRootHash.Bytes())
}

func StackTrieOldShaNewBlock(txns []*types.Transaction, expectedRoot string) {
	hasher := trie.NewStackTrie(nil)
	txnRootHash := OldDeriveSha(types.Transactions(txns), hasher)
	CheckHash(expectedRoot, txnRootHash.Bytes())
}

func StackTrieNewShaNewBlock(txns []*types.Transaction, expectedRoot string) {
	hasher := trie.NewStackTrie(nil)
	txnRootHash := DeriveSha(types.Transactions(txns), hasher)
	CheckHash(expectedRoot, txnRootHash.Bytes())
}

func SimpleTrieOldShaNewBlock(txns []*types.Transaction, expectedRoot string) {
	list := types.Transactions(txns)
	trie := simpletrie.NewTrie()

	InsertTrieIndexOrder(list, trie)
	CheckHash(expectedRoot, trie.Hash())
}

func SimpleTrieNewShaNewBlock(txns []*types.Transaction, expectedRoot string) {
	list := types.Transactions(txns)
	trie := simpletrie.NewTrie()

	InsertTrieByteOrder(list, trie)
	CheckHash(expectedRoot, trie.Hash())
}

func TestTrieHash(blockNum int, txnRoot string) {
	fmt.Println("Testing trie hash for block", blockNum)
	txns := TransactionsFromJSON(blockNum)
	txnRootBytes, err := hex.DecodeString(txnRoot[2:])
	PanicError(err)
	fmt.Println("Expected txn root:", txnRootBytes)

	fmt.Println("Old Trie structure, Old DeriveSha")
	TrieOldShaNewBlock(txns, txnRoot)

	fmt.Println("Old Trie structure, New DeriveSha")
	TrieNewShaNewBlock(txns, txnRoot)

	fmt.Println("StackTrie structure, Old DeriveSha")
	StackTrieOldShaNewBlock(txns, txnRoot)

	fmt.Println("StackTrie structure, New DeriveSha")
	StackTrieNewShaNewBlock(txns, txnRoot)

	fmt.Println("simpletrie structure, old DeriveSha")
	SimpleTrieOldShaNewBlock(txns, txnRoot)

	fmt.Println("simpletrie structure, new DeriveSha")
	SimpleTrieNewShaNewBlock(txns, txnRoot)
}

func main() {
	// TODO: prelondon blocks trigger "hashedNode" for StackTrie insert
	//TestTrieHash(PreLondonBlockNum, PreLondonTxnsRoot)

	TestTrieHash(PostLondonBlockNum, PostLondonBlockTxnsRoot)
}
