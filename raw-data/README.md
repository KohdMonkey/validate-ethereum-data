# Validating RLP-encoded data from Geth

Querying data from Geth through the standard RPC methods invokes calls to 
the eth namespace, and will return the data requested as a JSON object. We 
would like to pass this data to the enclave in its RLP-encoded form, but 
this would involve an extra encoding step for the process running outside 
the enclave. Geth provides a debug namespace that allows querying of data in 
its raw, rlp-encoded form, which we can use to eliminate the extra encoding 
step. This folder contains code that make the RPC calls to Geth to fetch 
this raw data and validates it. We will validate data from three RPC calls.

1. **debug_getBlockRlp**
2. **debug_getHeaderRlp**
3. **debug_getRawReceipts**


### **debug_getBlockRlp**
Returns the rlp-encoded block. 

curl "endpoint" -X POST -H "Content-Type: application/json" --data '
{"method":"debug_getHeaderRlp","params":[blocknum-in-decimal],"id":1,
"jsonrpc":"2.0"}'

### **debug_getHeaderRlp**
Returns the rlp-encoded header.

curl "endpoint" -X POST -H "Content-Type: application/json" --data '
{"method":"debug_getBlockRlp","params":[blocknum-in-decimal],"id":1,
"jsonrpc":"2.0"}'

### **debug_getRawReceipts**
Returns the list of receipts for a block in rlp format.

curl "endpoint" -X POST -H "Content-Type: application/json" --data '
{"method":"debug_getRawReceipts","params":[block-num-as-hex-string],"id":1,
"jsonrpc":"2.0"}'