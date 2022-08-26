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
	"net/http"
	"os"
	"strconv"
)

const (
	DataDir  = "data"
	BlockNum = 15209997

	ReceiptsRoot = "0x5ff308f613dd6b9cc880622fe638c4099c38fc85d02db7c738952618380360fd"

	HeaderHash = "0x868248867378bf14da3923ba2242e00a97154f390956ee5d5f7793f97920c047"
	// BlockHash A block's hash is just the hash of its header.
	BlockHash = "0x868248867378bf14da3923ba2242e00a97154f390956ee5d5f7793f97920c047"
)

func ReceiptsFromJSON() []*types.Receipt {
	receiptsFile := DataDir + "/block-" + strconv.Itoa(BlockNum) + "-receipts.json"
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

func HeaderFromJSON() types.Header {
	transactionsFile := DataDir + "/block-" + strconv.Itoa(BlockNum) + "-header.json"
	jsonFile, err := os.Open(transactionsFile)
	PanicError(err)
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	PanicError(err)

	var header types.Header
	json.Unmarshal(byteValue, &header)
	return header
}

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}

func ExitError(err string) {
	fmt.Println(err)
	os.Exit(1)
}

type RequestData struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

type ResponseData struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  string `json:"result"`
}

type ResponseDataArray struct {
	Jsonrpc string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Result  []string `json:"result"`
}

func ParseResponse(response *http.Response) ResponseData {
	var resData ResponseData
	json.NewDecoder(response.Body).Decode(&resData)

	return resData
}

func ParseResponseArray(response *http.Response) ResponseDataArray {
	var resData ResponseDataArray
	json.NewDecoder(response.Body).Decode(&resData)

	return resData
}

func ResultToByteArray(resString string) []byte {
	data, err := hex.DecodeString(resString[2:])
	PanicError(err)
	return data
}

func BytesToHeader(dataBytes []byte) *types.Header {
	var header *types.Header
	err := rlp.DecodeBytes(dataBytes, &header)
	PanicError(err)

	return header
}

func BytesToBlock(dataBytes []byte) *types.Block {
	var block *types.Block
	err := rlp.DecodeBytes(dataBytes, &block)
	PanicError(err)

	return block
}

func ExecuteRequest(data RequestData) *http.Response {
	payloadBytes, err := json.Marshal(data)
	PanicError(err)

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "http://127.0.0.1:8545/", body)
	PanicError(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	PanicError(err)

	return resp
}

func VerifyRawHeader() {
	fmt.Println("Verifying raw header... ")
	data := RequestData{
		Method:  "debug_getHeaderRlp",
		Params:  []interface{}{BlockNum},
		ID:      1,
		Jsonrpc: "2.0",
	}

	// fetch rlp-encoded header and parse response into bytes
	resp := ExecuteRequest(data)
	defer resp.Body.Close()

	resData := ParseResponse(resp)
	headerBytes := ResultToByteArray(resData.Result)

	//load header bytes from json file and check if rlp-encoded bytes match
	headerFromJson := HeaderFromJSON()
	headerBytesFromJson, err := rlp.EncodeToBytes(&headerFromJson)
	PanicError(err)

	if !bytes.Equal(headerBytesFromJson, headerBytes) {
		ExitError("raw header from json does not match raw header from rpc")
	}

	// construct header from raw bytes and calculate the hash
	header := BytesToHeader(headerBytes)
	headerHashStr := header.Hash().String()

	if headerHashStr != HeaderHash {
		ExitError("header hash does not match")
	}

	fmt.Println("header hash matches")
	fmt.Println("Hash from client: ", headerHashStr)
	fmt.Println("Expected hash: ", HeaderHash)
}

func VerifyRawBlock() {
	fmt.Println("Verifying raw block... ")
	data := RequestData{
		Method:  "debug_getBlockRlp",
		Params:  []interface{}{BlockNum},
		ID:      1,
		Jsonrpc: "2.0",
	}

	// fetch rlp-encoded block and parse response into bytes
	resp := ExecuteRequest(data)
	defer resp.Body.Close()

	// construct block from raw bytes and calculate the hash
	resData := ParseResponse(resp)
	blockBytes := ResultToByteArray(resData.Result)
	block := BytesToBlock(blockBytes)
	blockHashStr := block.Hash().String()

	if blockHashStr != BlockHash {
		ExitError("block hash does not match")
	}

	fmt.Println("block hash matches")
	fmt.Println("Hash from client: ", blockHashStr)
	fmt.Println("Expected hash: ", BlockHash)
}

func VerifyRawReceipts() {
	fmt.Println("Verifying raw receipts... ")
	BlockNumString := fmt.Sprintf("0x%x", BlockNum)
	data := RequestData{
		Method:  "debug_getRawReceipts",
		Params:  []interface{}{BlockNumString},
		ID:      1,
		Jsonrpc: "2.0",
	}
	// fetch rlp-encoded receipts and parse response into receipts
	resp := ExecuteRequest(data)
	defer resp.Body.Close()

	resData := ParseResponseArray(resp)
	encodedReceipts := resData.Result

	numReceipts := len(encodedReceipts)
	receipts := make([]*types.Receipt, numReceipts)
	receiptsBytesArr := make([][]byte, numReceipts)
	for i := 0; i < numReceipts; i++ {
		receipts[i] = new(types.Receipt)
		receiptHex := encodedReceipts[i][2:]
		receiptBytes := common.FromHex(receiptHex)
		receiptsBytesArr[i] = receiptBytes

		err := receipts[i].UnmarshalBinary(receiptBytes)
		if err != nil {
			fmt.Println("error unmarshalling receipt")
		}
	}

	// load receipts from json and check if receipts matches
	receiptsFromJson := ReceiptsFromJSON()
	for i := 0; i < numReceipts; i++ {
		receiptBinaryFromJson, err := receiptsFromJson[i].MarshalBinary()
		PanicError(err)

		if !bytes.Equal(receiptsBytesArr[i], receiptBinaryFromJson) {
			errStr := fmt.Sprintf(
				"receipt %d from json does not match receipt from rpc", i)
			ExitError(errStr)
		}
	}

	// construct trie using receipts and retrieve receipt root hash
	hasher := trie.NewStackTrie(nil)
	treeHash := types.DeriveSha(types.Receipts(receipts), hasher)
	treeHashStr := treeHash.String()

	if treeHashStr != ReceiptsRoot {
		ExitError("receipts root does not match")
	}

	fmt.Println("receipts root matches")
	fmt.Println("root from client: ", treeHashStr)
	fmt.Println("Expected root: ", ReceiptsRoot)
}

func main() {
	VerifyRawHeader()
	fmt.Println("----------------------------------------------------")
	VerifyRawBlock()
	fmt.Println("----------------------------------------------------")
	VerifyRawReceipts()
}
