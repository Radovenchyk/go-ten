package host

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/obscuronet/obscuro-playground/go/obscuronode/nodecommon"
	"github.com/obscuronet/obscuro-playground/go/obscuronode/obscuroclient"
)

// An in-memory implementation of `clientserver.Client` that speaks directly to the node.
type inMemObscuroClient struct {
	nodeID     common.Address
	obscuroAPI ObscuroAPI
}

func NewInMemObscuroClient(nodeID int64, host *Node) obscuroclient.Client {
	return &inMemObscuroClient{
		nodeID:     common.BigToAddress(big.NewInt(nodeID)),
		obscuroAPI: *NewObscuroAPI(host),
	}
}

func (c *inMemObscuroClient) ID() common.Address {
	return c.nodeID
}

// Call bypasses RPC, and invokes methods on the node directly.
func (c *inMemObscuroClient) Call(result interface{}, method string, args ...interface{}) error {
	switch method {
	case obscuroclient.RPCSendTransactionEncrypted:
		if len(args) != 1 {
			return fmt.Errorf(">   Agg%d: expected 1 arg to %s, got %d", c.nodeID, obscuroclient.RPCSendTransactionEncrypted, len(args))
		}
		tx, ok := args[0].(nodecommon.EncryptedTx)
		if !ok {
			return fmt.Errorf("arg to %s was not of expected type nodecommon.EncryptedTx", obscuroclient.RPCSendTransactionEncrypted)
		}

		c.obscuroAPI.SendTransactionEncrypted(tx)

	case obscuroclient.RPCGetCurrentBlockHeadHeight:
		*result.(*int64) = c.obscuroAPI.GetCurrentBlockHeadHeight()

	case obscuroclient.RPCGetCurrentRollupHead:
		*result.(**nodecommon.Header) = c.obscuroAPI.GetCurrentRollupHead()

	case obscuroclient.RPCGetRollupHeader:
		if len(args) != 1 {
			return fmt.Errorf("expected 1 arg to %s, got %d", obscuroclient.RPCGetRollupHeader, len(args))
		}
		hash, ok := args[0].(common.Hash)
		if !ok {
			return fmt.Errorf("arg to %s was not of expected type common.Hash", obscuroclient.RPCGetRollupHeader)
		}

		*result.(**nodecommon.Header) = c.obscuroAPI.GetRollupHeader(hash)

	case obscuroclient.RPCGetTransaction:
		if len(args) != 1 {
			return fmt.Errorf("expected 1 arg to %s, got %d", obscuroclient.RPCGetTransaction, len(args))
		}
		hash, ok := args[0].(common.Hash)
		if !ok {
			return fmt.Errorf("arg to %s was not of expected type common.Hash", obscuroclient.RPCGetTransaction)
		}

		*result.(**nodecommon.L2Tx) = c.obscuroAPI.GetTransaction(hash)

	case obscuroclient.RPCBalance:
		if len(args) != 1 {
			return fmt.Errorf("expected 1 arg to %s, got %d", obscuroclient.RPCBalance, len(args))
		}
		address, ok := args[0].(common.Address)
		if !ok {
			return fmt.Errorf("arg to %s was not of expected type common.Address", obscuroclient.RPCBalance)
		}

		*result.(*uint64) = c.obscuroAPI.Balance(address)

	case obscuroclient.RPCNonce:
		if len(args) != 1 {
			return fmt.Errorf("expected 1 arg to %s, got %d", obscuroclient.RPCNonce, len(args))
		}
		address, ok := args[0].(common.Address)
		if !ok {
			return fmt.Errorf("arg to %s was not of expected type common.Address", obscuroclient.RPCNonce)
		}

		*result.(*uint64) = c.obscuroAPI.Nonce(address)

	case obscuroclient.RPCStopHost:
		c.obscuroAPI.StopHost()

	default:
		return fmt.Errorf("RPC method %s is unknown", method)
	}

	return nil
}

func (c *inMemObscuroClient) Stop() {
	// There is no RPC connection to close.
}
