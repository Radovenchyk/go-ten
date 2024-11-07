package keymanager

import (
	"fmt"
	"os"
	"path/filepath"

	gethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ten-protocol/go-ten/go/common/log"
	"github.com/ten-protocol/go-ten/go/enclave/core/egoutils"
	"github.com/ten-protocol/go-ten/tools/walletextension/common"
)

const (
	dataDir           = "/data"
	encryptionKeyFile = "encryption.key"
)

// GetEncryptionKey returns encryption key for the database
// We have a few different scenarios and depending on scenario we try to get / generate the key.
// 1.) We want to use sqlite databse -> we dont need to do anything since sqlite does not need encryption and is usually running in dev environments / for testing
// 2.) If we want to use CosmosDB we need to have encryption key
//
//	 I) We need to check if the key is already sealed and unseal it if so
//	II) If there is a URL to exchange the key with another enclave we need to get the key from there and seal it on this enclave
//	III) If the key is not sealed and we don't have the URL to exchange the key we need to generate a new one and seal it
func GetEncryptionKey(config common.Config, logger gethlog.Logger) ([]byte, error) {
	fmt.Println("Getting encryption key")
	var encryptionKey []byte

	// check if we are using sqlite database
	if config.DBType == "sqlite" {
		return nil, nil
	}

	// Check if we have a sealed key and try to unseal it
	fmt.Println("Trying to unseal key")
	encryptionKey, found, err := tryUnsealKey(filepath.Join(dataDir, encryptionKeyFile), config.InsideEnclave)
	if err != nil {
		return nil, err
	}

	// If we found a sealed key we can return it
	if found {
		return encryptionKey, nil
	}

	// TODO: Handle the case where we have to exchange the key with another enclave

	// Generate a new key if we don't have a sealed one
	fmt.Println("Generating new encryption key")
	encryptionKey, err = common.GenerateRandomKey()
	if err != nil {
		logger.Crit("unable to generate random encryption key", log.ErrKey, err)
		return nil, err
	}
	fmt.Println("Generated new encryption key")

	fmt.Println("Sealing new encryption key")
	err = trySealKey(encryptionKey, filepath.Join(dataDir, encryptionKeyFile), config.InsideEnclave)
	if err != nil {
		logger.Crit("unable to seal encryption key", log.ErrKey, err)
		return nil, err
	}
	fmt.Println("Sealed new encryption key")
	return encryptionKey, nil
}

// tryUnsealKey attempts to unseal an encryption key from disk
// Returns (key, found, error)
func tryUnsealKey(keyPath string, isEnclave bool) ([]byte, bool, error) {
	// Only attempt unsealing if we're in an SGX enclave
	if !isEnclave {
		return nil, false, nil
	}

	// Read the key and unseal if possible
	data, err := egoutils.ReadAndUnseal(keyPath)
	if err != nil {
		return nil, false, err
	}

	return data, true, nil
}

// trySealKey attempts to seal an encryption key to disk
// Only seals if running in an SGX enclave
func trySealKey(key []byte, keyPath string, isEnclave bool) error {
	// Only attempt sealing if we're in an SGX enclave
	if !isEnclave {
		return nil
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0o644); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Seal and persist the key
	if err := egoutils.SealAndPersist(string(key), keyPath, true); err != nil {
		return fmt.Errorf("failed to seal and persist key: %w", err)
	}

	return nil
}
