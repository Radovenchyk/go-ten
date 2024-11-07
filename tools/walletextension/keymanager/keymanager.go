package keymanager

import (
	gethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ten-protocol/go-ten/go/common/log"
	"github.com/ten-protocol/go-ten/tools/walletextension/common"
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

	// check if we are using sqlite database
	if config.DBType == "sqlite" {
		return nil, nil
	}

	// Check if we have a sealed key and try to unseal it

	// Generate a new key if we don't have a sealed one
	encryptionKey, err := common.GenerateRandomKey()
	if err != nil {
		logger.Crit("unable to generate random encryption key", log.ErrKey, err)
		return nil, err
	}

	// TODO: Sealing the key

	return encryptionKey, nil
}
