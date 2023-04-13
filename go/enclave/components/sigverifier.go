package components

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/obscuronet/go-obscuro/go/enclave/db"
)

type SignatureValidator struct {
	sequencerID gethcommon.Address
	storage     db.Storage
}

func NewSignatureValidator(seqID gethcommon.Address, storage db.Storage) *SignatureValidator {
	return &SignatureValidator{
		sequencerID: seqID,
		storage:     storage,
	}
}

func (ov *SignatureValidator) CheckSequencerSignature(headerHash *gethcommon.Hash, aggregator *gethcommon.Address, sigR *big.Int, sigS *big.Int) error {
	// Batches and rollups should only be produced by the sequencer.
	// todo (#718) - sequencer identities should be retrieved from the L1 management contract
	if !bytes.Equal(aggregator.Bytes(), ov.sequencerID.Bytes()) {
		return fmt.Errorf("expected batch to be produced by sequencer %s, but was produced by %s", ov.sequencerID.Hex(), aggregator.Hex())
	}

	if sigR == nil || sigS == nil {
		return fmt.Errorf("missing signature on batch")
	}

	pubKey, err := ov.storage.FetchAttestedKey(*aggregator)
	if err != nil {
		return fmt.Errorf("could not retrieve attested key for aggregator %s. Cause: %w", aggregator, err)
	}

	if !ecdsa.Verify(pubKey, headerHash.Bytes(), sigR, sigS) {
		return fmt.Errorf("could not verify ECDSA signature")
	}
	return nil
}