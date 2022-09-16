package obsclient

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/obscuronet/go-obscuro/go/common"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

// utils for converting to RPC message format - mostly ported from geth client

// Formats a transaction for sending to the enclave
func encodeTx(tx *common.L2Tx) string {
	txBinary, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	// We convert the transaction binary to the form expected for sending transactions via RPC.
	txBinaryHex := gethcommon.Bytes2Hex(txBinary)

	return "0x" + txBinaryHex
}

// toBlockNumArg helper ensures a nil big int is converted into "latest"
func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}
