from ethereum.config import default_config
from eth_utils import (
    encode_hex,
    decode_hex,
    keccak,
)
from ethereum import utils
from ethereum.utils import (
    address,
    int256,
    hash32,
    trie_root
)
from ethereum import trie
import rlp
from rlp.sedes import (
    big_endian_int,
    Binary,
    binary,
)


class BlockHeader(rlp.Serializable):
    """A block header. Adapted from ethereum.block.BlockHeader

    If the block with this header exists as an instance of :class:`Block`, the
    connection can be made explicit by setting :attr:`BlockHeader.block`. Then,
    :attr:`BlockHeader.state_root`, :attr:`BlockHeader.tx_list_root` and
    :attr:`BlockHeader.receipts_root` always refer to the up-to-date value in
    the block instance.

    :ivar parent_hash: the 32 byte hash of the previous block
    :ivar uncles_hash: the 32 byte hash of the RLP encoded list of uncle
                       headers
    :ivar coinbase: the 20 byte coinbase address
    :ivar state_root: the root of the block's state trie
    :ivar transaction_root: the root of the block's transaction trie
    :ivar receipts_root: the root of the block's receipts trie
    :ivar bloom: Bloom filter for the logs of all transactions in the block
    :ivar difficulty: the block's difficulty
    :ivar number: the number of ancestors of this block (0 for the genesis
                  block)
    :ivar gas_limit: the block's gas limit
    :ivar gas_used: the total amount of gas used by all transactions in this
                    block
    :ivar timestamp: a UNIX timestamp
    :ivar extra_data: up to 1024 bytes of additional data
    :ivar mixhash: calculated using nonce to provide prevalidation of PoW
    :ivar nonce: a 32 byte nonce constituting a proof-of-work, or the empty
                 string as a placeholder
    :ivar baseFeePerGas: minimum price for inclusion of a transaction into 
                         this block
    """

    fields = [
        ('parent_hash', hash32),
        ('uncles_hash', hash32),
        ('coinbase', address),
        ('state_root', trie_root),
        ('transaction_root', trie_root),
        ('receipts_root', trie_root),
        ('bloom', int256),
        ('difficulty', big_endian_int),
        ('number', big_endian_int),
        ('gas_limit', big_endian_int),
        ('gas_used', big_endian_int),
        ('timestamp', big_endian_int),
        ('extra_data', binary),
        ('mix_hash', binary),
        ('nonce', Binary(8, allow_empty=True)),
        ('base_fee_per_gas', big_endian_int),
    ]

    def __init__(self,
                 parent_hash=default_config['GENESIS_PREVHASH'],
                 uncles_hash=utils.sha3rlp([]),
                 coinbase=default_config['GENESIS_COINBASE'],
                 state_root=trie.BLANK_ROOT,
                 transaction_root=trie.BLANK_ROOT,
                 receipts_root=trie.BLANK_ROOT,
                 bloom=0,
                 difficulty=default_config['GENESIS_DIFFICULTY'],
                 number=0,
                 gas_limit=default_config['GENESIS_GAS_LIMIT'],
                 gas_used=0,
                 timestamp=0,
                 extra_data=b'',
                 mix_hash=default_config['GENESIS_MIXHASH'],
                 nonce=b'',
                 base_fee_per_gas=0):
        # at the beginning of a method, locals() is a dict of all arguments
        fields = {k: v for k, v in locals().items() if k not in ['self', '__class__']}
        if len(fields['coinbase']) == 40:
            fields['coinbase'] = decode_hex(fields['coinbase'])
        assert len(fields['coinbase']) == 20
        self.block = None
        super(BlockHeader, self).__init__(**fields)

    def hash(self) -> bytes:
        return keccak(rlp.encode(self))

    @property
    def hex_hash(self):
        return encode_hex(self.hash)

    @property
    def mining_hash(self):

        # exclude mixhash and nonce
        fields2 = [
            (field, sedes) for field, sedes in BlockHeader._meta.fields
            if field not in ["mixhash", "nonce"]
        ]

        class BlockHeader2(rlp.Serializable):
            fields = fields2

        _self = BlockHeader2(**{f:getattr(self, f) for (f, sedes) in fields2})

        return utils.sha3(rlp.encode(
            _self, BlockHeader2))

    @property
    def signing_hash(self):

        # exclude extra_data
        fields3 = [
            (field, sedes) for field, sedes in BlockHeader._meta.fields
            if field not in ["extra_data"]
        ]

        class BlockHeader3(rlp.Serializable):
            fields = fields3

        _self = BlockHeader3(**{f:getattr(self, f) for (f, sedes) in fields3})

        return utils.sha3(rlp.encode(
            _self, BlockHeader3))

    def to_dict(self):
        """Serialize the header to a readable dictionary."""
        d = {}
        for field in ('prevhash', 'uncles_hash', 'extra_data', 'nonce',
                      'mixhash'):
            d[field] = '0x' + encode_hex(getattr(self, field))
        for field in ('state_root', 'tx_list_root', 'receipts_root',
                      'coinbase'):
            d[field] = encode_hex(getattr(self, field))
        for field in ('number', 'difficulty', 'gas_limit', 'gas_used',
                      'timestamp'):
            d[field] = utils.to_string(getattr(self, field))
        d['bloom'] = encode_hex(int256.serialize(self.bloom))
        assert len(d) == len(BlockHeader.fields)
        return d

    def __repr__(self):
        return '<%s(#%d %s)>' % (self.__class__.__name__, self.number,
                                 encode_hex(self.hash)[:8])

    def __eq__(self, other):
        """Two blockheader are equal iff they have the same hash."""
        return isinstance(other, BlockHeader) and self.hash == other.hash

    def __hash__(self):
        return utils.big_endian_to_int(self.hash)

    def __ne__(self, other):
        return not self.__eq__(other)