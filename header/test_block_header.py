import block_header

from eth_utils import to_bytes, to_hex

import json

BLOCK_FILE = 'block.json'

def load_block_from_file():
    with open(BLOCK_FILE, "r") as f:
        block_json = json.load(f)
    return block_json

def hex_str_to_binary(hex_str: str) -> bytes:
    hex_str = hex_str[2:]
    return bytes.fromhex(hex_str)

def test_block_hash():
    block = load_block_from_file()
    block_dict = block['result']

    header = block_header.BlockHeader(
        parent_hash = hex_str_to_binary(block_dict['parentHash']),
        uncles_hash = hex_str_to_binary(block_dict['sha3Uncles']),
        coinbase = hex_str_to_binary(block_dict['miner']),
        state_root = hex_str_to_binary(block_dict['stateRoot']),
        transaction_root = hex_str_to_binary(block_dict['transactionsRoot']),
        receipts_root = hex_str_to_binary(block_dict['receiptsRoot']),
        bloom = int(block_dict['logsBloom'], 0),
        difficulty = int(block_dict['difficulty'], 0),
        number = int(block_dict['number'], 0),
        gas_limit = int(block_dict['gasLimit'], 0),
        gas_used = int(block_dict['gasUsed'], 0),
        timestamp = int(block_dict['timestamp'], 0),
        extra_data = hex_str_to_binary(block_dict['extraData']),
        mix_hash = hex_str_to_binary(block_dict['mixHash']),
        nonce = hex_str_to_binary(block_dict['nonce']),
        base_fee_per_gas = int(block_dict['baseFeePerGas'], 0)
    )

    assert to_hex(header.hash()) == block_dict['hash'] , "header hash doesn't match!"

    
test_block_hash()