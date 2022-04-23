package l1client

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/obscuronet/obscuro-playground/go/l1client/txhandler"
	"github.com/obscuronet/obscuro-playground/go/l1client/wallet"
	"github.com/obscuronet/obscuro-playground/go/log"
	"github.com/obscuronet/obscuro-playground/go/obscurocommon"
)

var (
	connectionTimeout = 15 * time.Second
	nonceLock         sync.RWMutex
)

type EthNode struct {
	client          *ethclient.Client
	id              common.Address // TODO remove the id common.Address
	wallet          wallet.Wallet
	chainID         int
	txHandler       txhandler.TxHandler
	contractAddress common.Address
}

// NewEthClient instantiates a new l1client.Client that connects to an ethereum node
func NewEthClient(id common.Address, ipaddress string, port uint, wallet wallet.Wallet, contractAddress common.Address) (Client, error) {
	client, err := connect(ipaddress, port)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to the eth node - %w", err)
	}

	log.Log(fmt.Sprintf("Initializing eth node at contract: %s", contractAddress))
	return &EthNode{
		client:          client,
		id:              id,
		wallet:          wallet, // TODO this does not need to be coupled together
		chainID:         1337,
		txHandler:       txhandler.NewEthTxHandler(contractAddress),
		contractAddress: contractAddress,
	}, nil
}

func (e *EthNode) FetchHeadBlock() (*types.Block, uint64) {
	blk, err := e.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	return blk, blk.Number().Uint64()
}

func (e *EthNode) Info() Info {
	return Info{
		ID: e.id,
	}
}

func (e *EthNode) BlocksBetween(startingBlock *types.Block, lastBlock *types.Block) []*types.Block {
	// TODO this should be a stream
	var blocksBetween []*types.Block
	var err error

	for currentBlk := lastBlock; currentBlk != nil && currentBlk.Hash() != startingBlock.Hash() && currentBlk.ParentHash() != common.HexToHash(""); {
		currentBlk, err = e.FetchBlock(currentBlk.ParentHash())
		if err != nil {
			panic(err)
		}
		blocksBetween = append(blocksBetween, currentBlk)
	}

	return blocksBetween
}

func (e *EthNode) IsBlockAncestor(block *types.Block, maybeAncestor obscurocommon.L1RootHash) bool {
	if maybeAncestor == block.Hash() || maybeAncestor == obscurocommon.GenesisBlock.Hash() {
		return true
	}

	if block.Number().Int64() == int64(obscurocommon.L1GenesisHeight) {
		return false
	}

	resolvedBlock, err := e.FetchBlock(maybeAncestor)
	if err != nil {
		panic(err)
	}
	if resolvedBlock == nil {
		if resolvedBlock.Number().Int64() >= block.Number().Int64() {
			return false
		}
	}

	p, err := e.FetchBlock(block.ParentHash())
	if err != nil {
		panic(err)
	}
	if p == nil {
		return false
	}

	return e.IsBlockAncestor(p, maybeAncestor)
}

func (e *EthNode) RPCBlockchainFeed() []*types.Block {
	var availBlocks []*types.Block

	block, err := e.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	availBlocks = append(availBlocks, block)

	for {
		// todo set this to genesis hash
		if block.ParentHash().Hex() == "0x0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}

		block, err = e.client.BlockByHash(context.Background(), block.ParentHash())
		if err != nil {
			panic(err)
		}

		availBlocks = append(availBlocks, block)
	}

	// TODO double check the list is ordered [genesis, 1, 2, 3, 4, ..., last]
	// TODO It's pretty ugly but it avoids creating a new slice
	// TODO The approach of feeding all the blocks should change from all-blocks-in-memory to a stream
	for i, j := 0, len(availBlocks)-1; i < j; i, j = i+1, j-1 {
		availBlocks[i], availBlocks[j] = availBlocks[j], availBlocks[i]
	}
	return availBlocks
}

func (e *EthNode) IssueCustomTx(tx types.TxData) (*types.Transaction, error) {
	signedTx, err := e.wallet.SignTransaction(e.chainID, tx)
	if err != nil {
		panic(err)
	}

	return signedTx, e.client.SendTransaction(context.Background(), signedTx)
}

func (e *EthNode) TransactionReceipt(hash common.Hash) (*types.Receipt, error) {
	return e.client.TransactionReceipt(context.Background(), hash)
}

func (e *EthNode) BroadcastTx(tx *obscurocommon.L1TxData) {
	nonceLock.Lock()
	defer nonceLock.Unlock()

	fromAddr := e.wallet.Address()
	nonce, err := e.client.PendingNonceAt(context.Background(), fromAddr)
	if err != nil {
		panic(err)
	}

	formattedTx, err := e.txHandler.PackTx(tx, fromAddr, nonce)
	if err != nil {
		panic(err)
	}

	_, err = e.IssueCustomTx(formattedTx)
	if err != nil {
		panic(err)
	}
}

func (e *EthNode) BlockListener() chan *types.Header {
	ch := make(chan *types.Header, 1)
	subs, err := e.client.SubscribeNewHead(context.Background(), ch)
	if err != nil {
		panic(err)
	}
	// we should hook the subs to cleanup
	fmt.Println(subs)

	return ch
}

func (e *EthNode) FetchBlockByNumber(n *big.Int) (*types.Block, error) {
	return e.client.BlockByNumber(context.Background(), n)
}

func (e *EthNode) FetchBlock(hash common.Hash) (*types.Block, error) {
	return e.client.BlockByHash(context.Background(), hash)
}

func connect(ipaddress string, port uint) (*ethclient.Client, error) {
	var err error
	var c *ethclient.Client
	for start := time.Now(); time.Since(start) < connectionTimeout; time.Sleep(time.Second) {
		c, err = ethclient.Dial(fmt.Sprintf("ws://%s:%d", ipaddress, port))
		if err == nil {
			break
		}
	}

	return c, err
}
