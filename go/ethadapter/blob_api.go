package ethadapter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"reflect"
	"strconv"
)

type BlobSidecar struct {
	Blob          Blob         `json:"blob"`
	Index         Uint64String `json:"index"`
	KZGCommitment Bytes48      `json:"kzg_commitment"`
	KZGProof      Bytes48      `json:"kzg_proof"`
}

type APIBlobSidecar struct {
	Index             Uint64String            `json:"index"`
	Blob              Blob                    `json:"blob"`
	KZGCommitment     Bytes48                 `json:"kzg_commitment"`
	KZGProof          Bytes48                 `json:"kzg_proof"`
	SignedBlockHeader SignedBeaconBlockHeader `json:"signed_block_header"`
	// The inclusion-proof of the blob-sidecar into the beacon-block is ignored,
	// since we verify blobs by their versioned hashes against the execution-layer block instead.
}

func (sc *APIBlobSidecar) BlobSidecar() *BlobSidecar {
	return &BlobSidecar{
		Blob:          sc.Blob,
		Index:         sc.Index,
		KZGCommitment: sc.KZGCommitment,
		KZGProof:      sc.KZGProof,
	}
}

type SignedBeaconBlockHeader struct {
	Message BeaconBlockHeader `json:"message"`
	// signature is ignored, since we verify blobs against EL versioned-hashes
}

type BeaconBlockHeader struct {
	Slot          Uint64String `json:"slot"`
	ProposerIndex Uint64String `json:"proposer_index"`
	ParentRoot    Bytes32      `json:"parent_root"`
	StateRoot     Bytes32      `json:"state_root"`
	BodyRoot      Bytes32      `json:"body_root"`
}

type APIGetBlobSidecarsResponse struct {
	Data []*APIBlobSidecar `json:"data"`
}

type ReducedGenesisData struct {
	GenesisTime Uint64String `json:"genesis_time"`
}

type APIGenesisResponse struct {
	Data ReducedGenesisData `json:"data"`
}

type ReducedConfigData struct {
	SecondsPerSlot Uint64String `json:"SECONDS_PER_SLOT"`
}

type APIConfigResponse struct {
	Data ReducedConfigData `json:"data"`
}

type APIVersionResponse struct {
	Data VersionInformation `json:"data"`
}

type VersionInformation struct {
	Version string `json:"version"`
}

// Uint64String is a decimal string representation of an uint64, for usage in the Beacon API JSON encoding
type Uint64String uint64

func (v Uint64String) MarshalText() (out []byte, err error) {
	out = strconv.AppendUint(out, uint64(v), 10)
	return
}

func (v *Uint64String) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 0, 64)
	if err != nil {
		return err
	}
	*v = Uint64String(n)
	return nil
}

type Bytes48 [48]byte

func (b *Bytes48) UnmarshalJSON(text []byte) error {
	return hexutil.UnmarshalFixedJSON(reflect.TypeOf(b), text, b[:])
}

func (b *Bytes48) UnmarshalText(text []byte) error {
	return hexutil.UnmarshalFixedText("Bytes32", text, b[:])
}

func (b Bytes48) MarshalText() ([]byte, error) {
	return hexutil.Bytes(b[:]).MarshalText()
}

func (b Bytes48) String() string {
	return hexutil.Encode(b[:])
}

// TerminalString implements log.TerminalStringer, formatting a string for console
// output during logging.
func (b Bytes48) TerminalString() string {
	return fmt.Sprintf("%x..%x", b[:3], b[45:])
}

type Bytes32 [32]byte

func (b *Bytes32) UnmarshalJSON(text []byte) error {
	return hexutil.UnmarshalFixedJSON(reflect.TypeOf(b), text, b[:])
}

func (b *Bytes32) UnmarshalText(text []byte) error {
	return hexutil.UnmarshalFixedText("Bytes32", text, b[:])
}

func (b Bytes32) MarshalText() ([]byte, error) {
	return hexutil.Bytes(b[:]).MarshalText()
}

func (b Bytes32) String() string {
	return hexutil.Encode(b[:])
}
