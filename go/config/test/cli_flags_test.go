package config

import (
	"flag"
	"github.com/ten-protocol/go-ten/go/common"
	"github.com/ten-protocol/go-ten/go/config"
	enclavecontainer "github.com/ten-protocol/go-ten/go/enclave/container"
	hostcontainer "github.com/ten-protocol/go-ten/go/host/container"
	"os"
	"testing"
)

const defaultHost = "/../templates/default_host_config.yaml"
const defaultEnclave = "/../templates/default_enclave_config.yaml"
const overrideConfig = "/partial.yaml"

// Same mechanism for host and enclave
func TestHostConfigIsParsedFromYamlFileIfConfigFlagIsPresent(t *testing.T) {
	resetFlagSet()

	l1WebsocketURL := "ws://0.0.0.0:8546"
	logLevel := 3

	// Back up the original os.Args to be available after unit test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Mock os.Args
	os.Args = []string{"your-program", "-config", wd + defaultHost}

	rParams, _, err := config.LoadFlagStrings(config.Host)
	cfg, err := hostcontainer.ParseConfig(rParams)
	if err != nil {
		t.Fatalf("could not parse config. Cause: %s", err)
	}
	if cfg.L1WebsocketURL != l1WebsocketURL || cfg.LogLevel != logLevel {
		t.Fatalf("config file was not parsed from YAML. Expected l1WebsockerURL of %s"+
			"and logLevel %d, got %s and %d", l1WebsocketURL, logLevel, cfg.L1WebsocketURL, cfg.LogLevel)
	}
}

// The default config will set the regular values including logLevel 3, override will swap logLevel
func TestEnclaveOverrideAdditiveReplacementOfDefaultConfig(t *testing.T) {
	resetFlagSet()

	gasBatchExecutionLimit := uint64(300_000_000_000)
	logLevel := 2

	// Back up the original os.Args to be available after unit test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Mock os.Args
	os.Args = []string{
		"your-program",
		"-config", wd + defaultEnclave,
		"-override", wd + overrideConfig,
	}

	rParams, _, err := config.LoadFlagStrings(config.Enclave)
	cfg, err := enclavecontainer.ParseConfig(rParams)
	if err != nil {
		t.Fatalf("could not parse config. Cause: %s", err)
	}
	if cfg.GasBatchExecutionLimit != gasBatchExecutionLimit {
		t.Fatalf("config file was not parsed from YAML. Expected gasBatchExecutionLimit of %d, got %d", gasBatchExecutionLimit, cfg.GasBatchExecutionLimit)
	}
	if cfg.LogLevel != logLevel {
		t.Fatalf("override failed, logLevel of default was 3 but should have overriden to %d", logLevel)
	}
}

func TestHostFlagOverridesDefaultProperty(t *testing.T) {
	resetFlagSet()

	// Back up the original os.Args to be available after unit test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Args = []string{
		"your-program",
		"-config", wd + defaultHost,
		"-nodeType", "sequencer",
	}

	rParams, _, err := config.LoadFlagStrings(config.Host)
	cfg, err := hostcontainer.ParseConfig(rParams)
	if err != nil {
		t.Fatalf("could not parse config. Cause: %s", err)
	}

	// Assert that the flag value overrides the default configuration
	if cfg.NodeType != common.Sequencer {
		t.Fatalf("default config was not loaded. Expected nodeType of %s, got %s", common.Validator, cfg.NodeType)
	}
}

func TestEnclaveEnvVarOverridesDefaultConfigAndFlag(t *testing.T) {
	resetFlagSet()

	// Back up the original os.Args to be available after unit test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set environment variable which will override below flag for nodeType
	err := os.Setenv("EDGELESSDBHOST", "testHost") // default is ""
	if err != nil {
		t.Fatalf("could not set environment variable. Cause: %s", err)
	}
	err = os.Setenv("NODETYPE", "validator")
	if err != nil {
		t.Fatalf("could not set environment variable. Cause: %s", err)
	}
	err = os.Setenv("LOGLEVEL", "2")
	if err != nil {
		t.Fatalf("could not set environment variable. Cause: %s", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Args = []string{
		"your-program",
		"-config", wd + defaultEnclave,
		"-nodeType", "sequencer", // flag will be overridden by env var (flag in space format)
		"-logLevel=1",           // flag will be overridden by env var (flag in = format)
		"-logPath", "/tmp/logs", // keep override config because no envVar
	}

	rParams, _, err := config.LoadFlagStrings(config.Enclave)
	cfg, err := enclavecontainer.ParseConfig(rParams)
	if err != nil {
		t.Fatalf("could not parse config. Cause: %s", err)
	}

	// Assert that the flag value overrides the default configuration
	if cfg.EdgelessDBHost != "testHost" {
		t.Fatalf("env override not successful. Expected edgelessDbHost of %s, got %s", "testHost", cfg.EdgelessDBHost)
	}
	if cfg.NodeType != common.Validator {
		t.Fatalf("env override not successful. Expected nodeType of %s, got %s", common.Validator, cfg.NodeType)
	}
	if cfg.LogLevel != 2 {
		t.Fatalf("env override not successful. Expected logLevel of %d, got %d", 2, cfg.LogLevel)
	}
	if cfg.LogPath != "/tmp/logs" {
		t.Fatalf("flag override failed. Expected logPath of %s, got %s", "/tmp/logs", cfg.LogPath)
	}
}

func TestEnclaveConfigJson(t *testing.T) {
	resetFlagSet()

	confEIC := config.EnclaveInputConfig{
		HostID:                    "example-host-id",
		HostAddress:               "example-host-address",
		Address:                   "example-address",
		NodeType:                  "example-node-type",
		L1ChainID:                 123,
		TenChainID:                456,
		WillAttest:                true,
		ValidateL1Blocks:          false,
		GenesisJSON:               "{}",
		ManagementContractAddress: "example-management-contract-address",
		LogLevel:                  1,
		LogPath:                   "example-log-path",
		UseInMemoryDB:             true,
		EdgelessDBHost:            "example-edgeless-db-host",
		SqliteDBPath:              "example-sqlite-db-path",
		ProfilerEnabled:           true,
		MinGasPrice:               1000,
		MessageBusAddress:         "example-message-bus-address",
		SequencerP2PAddress:       "example-sequencer-p2p-address",
		TenGenesis:                "example-ten-genesis",
		DebugNamespaceEnabled:     true,
		MaxBatchSize:              2000,
		MaxRollupSize:             3000,
		GasPaymentAddress:         "example-gas-payment-address",
		BaseFee:                   500,
		GasBatchExecutionLimit:    6000,
		GasLocalExecutionCap:      7000,
		RPCTimeout:                10,
	}

	err := confEIC.ToEnclaveConfigJson("./enclave.json")
	if err != nil {
		return
	}
}

// needed for subsequent runs testing flags
func resetFlagSet() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}
