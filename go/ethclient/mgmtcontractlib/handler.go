package mgmtcontractlib

import (
	"fmt"
	"math/big"

	"github.com/obscuronet/obscuro-playground/contracts"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/obscuronet/obscuro-playground/go/log"
	"github.com/obscuronet/obscuro-playground/go/obscurocommon"
	"github.com/obscuronet/obscuro-playground/go/obscuronode/nodecommon"
)

const methodBytesLen = 4

var (
	// TODO review estimating gas - these should not be static values
	defaultGasPrice = big.NewInt(20000000000)
	defaultGas      = uint64(1024_000_000)
)

type TxHandler interface {
	// PackTx receives an obscurocommon.L1TxData object and packs it into a types.TxData object
	// Nonce generation, transaction signature and any other operations are responsibility of the caller
	PackTx(tx *obscurocommon.L1TxData, from common.Address, nonce uint64) (types.TxData, error)

	// UnPackTx receives a *types.Transaction and converts it to an obscurocommon.L1TxData pointer
	// Any transaction that is not calling the management contract is purposefully ignored
	UnPackTx(tx *types.Transaction) *obscurocommon.L1TxData
}

type EthTxHandler struct {
	contractAddr common.Address
}

func NewEthTxHandler(contractAddress common.Address) TxHandler {
	return &EthTxHandler{
		contractAddr: contractAddress,
	}
}

func (h *EthTxHandler) PackTx(tx *obscurocommon.L1TxData, fromAddr common.Address, nonce uint64) (types.TxData, error) {
	ethTx := &types.LegacyTx{
		Nonce:    nonce,
		GasPrice: defaultGasPrice,
		Gas:      defaultGas,
		To:       &h.contractAddr,
	}

	// TODO each of these cases should be a function:
	// TODO like: func createRollupTx() or func createDepositTx()
	// TODO And then eventually, these functions would be called directly, when we get rid of our special format. (we'll have to change the mock thing as well for that)
	switch tx.TxType {
	case obscurocommon.DepositTx:
		ethTx.Value = big.NewInt(int64(tx.Amount))
		data, err := contracts.MgmtContractABIJSON.Pack("Deposit", tx.Dest)
		if err != nil {
			panic(err)
		}
		ethTx.Data = data
		log.Log(fmt.Sprintf("Broadcasting - Issuing DepositTx - Addr: %s deposited %d to %s ",
			fromAddr, tx.Amount, tx.Dest))

	case obscurocommon.RollupTx:
		r, err := nodecommon.DecodeRollup(tx.Rollup)
		if err != nil {
			panic(err)
		}
		zipped, err := Compress(tx.Rollup)
		if err != nil {
			panic(err)
		}
		encRollupData := EncodeToString(zipped)
		data, err := contracts.MgmtContractABIJSON.Pack("AddRollup", encRollupData)
		if err != nil {
			panic(err)
		}

		ethTx.Data = data
		log.Log(fmt.Sprintf("Broadcasting - Issuing Rollup: %s - %d txs - datasize: %d - gas: %d \n", r.Hash(), len(r.Transactions), len(data), ethTx.Gas))

	case obscurocommon.StoreSecretTx:
		data, err := contracts.MgmtContractABIJSON.Pack("StoreSecret", EncodeToString(tx.Secret))
		if err != nil {
			panic(err)
		}
		ethTx.Data = data
		log.Log(fmt.Sprintf("Broadcasting - Issuing StoreSecretTx: encoded as %s", EncodeToString(tx.Secret)))
	case obscurocommon.RequestSecretTx:
		data, err := contracts.MgmtContractABIJSON.Pack("RequestSecret")
		if err != nil {
			panic(err)
		}
		ethTx.Data = data
		log.Log("Broadcasting - Issuing RequestSecret")
	}

	return ethTx, nil
}

func (h *EthTxHandler) UnPackTx(tx *types.Transaction) *obscurocommon.L1TxData {
	// ignore transactions that are not calling the contract
	if tx.To() == nil || tx.To().Hex() != h.contractAddr.Hex() || len(tx.Data()) == 0 {
		log.Log(fmt.Sprintf("UnpackTx: Ignoring transaction %+v", tx))
		return nil
	}

	method, err := contracts.MgmtContractABIJSON.MethodById(tx.Data()[:methodBytesLen])
	if err != nil {
		panic(err)
	}

	l1txData := obscurocommon.L1TxData{
		TxType:      0,
		Attestation: obscurocommon.AttestationReport{},
		Amount:      0,
		Dest:        common.Address{},
	}
	contractCallData := map[string]interface{}{}
	switch method.Name {
	case contracts.DepositMethod:
		if err := method.Inputs.UnpackIntoMap(contractCallData, tx.Data()[4:]); err != nil {
			panic(err)
		}
		callData, found := contractCallData["dest"]
		if !found {
			panic("call data not found for dest")
		}

		l1txData.TxType = obscurocommon.DepositTx
		l1txData.Amount = tx.Value().Uint64()
		l1txData.Dest = callData.(common.Address)

	case contracts.AddRollupMethod:
		if err := method.Inputs.UnpackIntoMap(contractCallData, tx.Data()[4:]); err != nil {
			panic(err)
		}
		callData, found := contractCallData["rollupData"]
		if !found {
			panic("call data not found for rollupData")
		}
		zipped := DecodeFromString(callData.(string))
		l1txData.Rollup, err = Decompress(zipped)
		if err != nil {
			panic(err)
		}
		l1txData.TxType = obscurocommon.RollupTx

	case contracts.StoreSecretMethod:
		if err := method.Inputs.UnpackIntoMap(contractCallData, tx.Data()[4:]); err != nil {
			panic(err)
		}
		callData, found := contractCallData["inputSecret"]
		if !found {
			panic("call data not found for inputSecret")
		}
		l1txData.Secret = DecodeFromString(callData.(string))
		l1txData.TxType = obscurocommon.StoreSecretTx

	case contracts.RequestSecretMethod:
		l1txData.TxType = obscurocommon.RequestSecretTx
	}

	return &l1txData
}
