package node

import (
	"fmt"
	"github.com/sanity-io/litter"
	"github.com/ten-protocol/go-ten/go/common/docker"
	"github.com/ten-protocol/go-ten/go/config"
	"net"
	"os"
	"strconv"
)

var (
	Action          = "action"
	_hostDataDir    = "/data"        // this is how the directory is referenced within the host container
	_enclaveDataDir = "/enclavedata" // this is how the directory is references within the enclave container
)

type DockerNode struct {
	Action   string
	Cfg      *config.NodeConfig
	CliFlags config.CliFlagStringSet
	DryRun   bool
}

// DockerStartBundle is a struct that holds the configuration settings for a docker container
// ConfigParam whether cmd should expect a config replacement
// OverrideParam whether cmd should expect a config override (additive)
// See go/config/templates for the default configurations
type DockerStartBundle struct {
	Service                string
	DryRun                 bool
	ContainerName, Image   string
	Cmds                   []string
	Ports                  []int
	Envs, Devices, Volumes map[string]string
}

func (d *DockerStartBundle) Print() {
	fmt.Println("Configuration Settings for ", d.Service, ":")
	fmt.Println(litter.Sdump(*d))
}

func NewDockerNode(runParams config.RunParams, cfg *config.NodeConfig, flags config.CliFlagStringSet) *DockerNode {
	return &DockerNode{
		Action:   runParams[Action],
		Cfg:      cfg,
		CliFlags: flags,
		DryRun:   runParams[config.DryRunFlag] == "true",
	}
}

func (d *DockerNode) Start() error {
	// todo (@pedro) - this should probably be removed in the future
	if d.DryRun {
		fmt.Printf("Dry run mode enabled, processing configuration without starting container: \n%s\n\n", litter.Sdump(*d))
	} else {
		fmt.Printf("Starting Node %s with config: \n%s\n\n", d.Cfg.NodeDetails.NodeName, litter.Sdump(*d))
	}

	err := d.startEdgelessDB()
	if err != nil {
		return err
	}

	err = d.startEnclave()
	if err != nil {
		return err
	}

	err = d.startHost()
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerNode) Stop() error {
	fmt.Println("Stopping existing host and enclave")
	err := docker.StopAndRemove(d.Cfg.NodeDetails.NodeName + "-host")
	if err != nil {
		return err
	}

	err = docker.StopAndRemove(d.Cfg.NodeDetails.NodeName + "-enclave")
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerNode) Upgrade(networkCfg *config.NetworkInputConfig) error {
	// TODO this should probably be removed in the future
	fmt.Printf("Upgrading node %s with config: %+v\n", d.Cfg.NodeDetails.NodeName, d)

	err := d.Stop()
	if err != nil {
		return err
	}

	// Adjusts network params to the persisted if not matching current config
	d.Cfg.SetNetwork(networkCfg)

	fmt.Println("Starting upgraded host and enclave")
	err = d.startEnclave()
	if err != nil {
		return err
	}

	err = d.startHost()
	if err != nil {
		return err
	}

	return nil
}

func (d *DockerNode) startHost() error {
	envs := map[string]string{}
	cmd := []string{
		"/home/obscuro/go-obscuro/go/host/main/main",
	}

	if !d.Cfg.HostConfig.UseInMemoryDB {
		if d.Cfg.HostConfig.PostgresDBHost == "" {
			panic("postgresDBHost required when useInMemoryDB is false")
		}
	}

	exposedPorts := []int{
		int(d.Cfg.HostConfig.ClientRPCPortHTTP),
		int(d.Cfg.HostConfig.ClientRPCPortWS),
		getPort(d.Cfg.HostConfig.P2PPublicAddress), // p2pBindAddress / hostP2PPort,
	}

	hostVolume := map[string]string{d.Cfg.NodeDetails.NodeName + "-host-volume": _hostDataDir}

	// prepend config override to the command
	cmd = append([]string{"/home/obscuro/go-obscuro/go/host/main/config-entrypoint.sh"}, cmd...)

	// set the override config
	err := config.WriteConfigToEnv(d.Cfg.HostConfig, envs, config.OverrideEnvKey)
	if err != nil {
		return err
	}
	cmd = append(cmd, "-override", config.OverrideEnvKey+".yaml")

	dsb := &DockerStartBundle{
		config.Host.String(),
		d.DryRun,
		d.Cfg.NodeDetails.NodeName + "-host",
		d.Cfg.NodeImages.HostImage,
		cmd,
		exposedPorts,
		envs,
		nil,
		hostVolume,
	}

	return dsb.startOrReportDryRun()
}

func (d *DockerNode) startEnclave() error {
	devices := map[string]string{}
	var exposedPorts []int
	envs := map[string]string{
		"OE_SIMULATION": "1",
	}

	// default start of the enclave
	cmd := []string{
		"ego", "run", "/home/obscuro/go-obscuro/go/enclave/main/main",
	}

	if d.Cfg.NodeSettings.EnclaveDebug {
		cmd = []string{
			"dlv",
			"--listen=:2345",
			"--headless=true",
			"--log=true",
			"--api-version=2",
			"debug",
			"/home/obscuro/go-obscuro/go/enclave/main",
			"--",
		}
		exposedPorts = append(exposedPorts, 2345)
	}

	if d.Cfg.NodeSettings.IsSGXEnabled {
		devices["/dev/sgx_enclave"] = "/dev/sgx_enclave"
		devices["/dev/sgx_provision"] = "/dev/sgx_provision"

		envs["OE_SIMULATION"] = "0"

		// prepend the entry.sh execution
		cmd = append([]string{"/home/obscuro/go-obscuro/go/enclave/main/entry.sh"}, cmd...)
		cmd = append(cmd,
			"-edgelessDBHost", d.Cfg.NodeDetails.NodeName+"-edgelessdb",
			"-willAttest=true",
		)
	} else {
		cmd = append(cmd,
			"-sqliteDBPath", "/data/sqlite.db",
		)
	}

	// prepend config override to the command
	cmd = append([]string{"/home/obscuro/go-obscuro/go/enclave/main/config-entrypoint.sh"}, cmd...)

	// set the override config
	err := config.WriteConfigToEnv(d.Cfg.EnclaveConfig, envs, config.OverrideEnvKey)
	if err != nil {
		return err
	}
	cmd = append(cmd, "-override", config.OverrideEnvKey+".yaml")

	enclaveVolume := map[string]string{d.Cfg.NodeDetails.NodeName + "-enclave-volume": _enclaveDataDir}

	dsb := &DockerStartBundle{
		config.Enclave.String(),
		d.DryRun,
		d.Cfg.NodeDetails.NodeName + "-enclave",
		d.Cfg.NodeImages.EnclaveImage,
		cmd,
		exposedPorts,
		envs,
		devices,
		enclaveVolume,
	}

	return dsb.startOrReportDryRun()
}

func (d *DockerNode) startEdgelessDB() error {
	if !d.Cfg.NodeSettings.IsSGXEnabled {
		// Non-SGX hardware use sqlite database so EdgelessDB is not required.
		return nil
	}

	envs := map[string]string{
		"EDG_EDB_CERT_DNS": d.Cfg.NodeDetails.NodeName + "-edgelessdb",
	}

	devices := map[string]string{
		"/dev/sgx_enclave":   "/dev/sgx_enclave",
		"/dev/sgx_provision": "/dev/sgx_provision",
	}

	// only set the pccsAddr env var if it's defined
	if d.Cfg.NodeSettings.PccsAddr != "" {
		envs["PCCS_ADDR"] = d.Cfg.NodeSettings.PccsAddr
	}

	dsb := &DockerStartBundle{
		"edgelessdb",
		d.DryRun,
		d.Cfg.NodeDetails.NodeName + "-edgelessdb",
		d.Cfg.NodeImages.EdgelessDBImage,
		nil,
		nil,
		envs,
		devices,
		nil,
	}

	return dsb.startOrReportDryRun()
}

func (d *DockerStartBundle) startOrReportDryRun() error {
	if d.DryRun {
		d.Print()
		return nil
	}
	_, err := docker.StartNewContainer(d.ContainerName, d.Image, d.Cmds, d.Ports, d.Envs, d.Devices, d.Volumes)
	return err
}

// appendConfigStaticFlagEnvOverrides takes in an envs map and applies layered override based on the
// configurations in file < program flags < environment variables
func (d *DockerNode) appendConfigStaticFlagEnvOverrides(t config.TypeConfig, envs map[string]string) map[string]string {
	// configuration properties derived as env vars
	envs = config.MergeEnvMaps(envs, d.Cfg.GetConfigAsEnvVars(t))
	// override with any program flags
	envs = config.MergeEnvMaps(envs, d.CliFlags)
	// Override with any explicit env variables
	for key := range envs {
		if val, exists := os.LookupEnv(key); exists {
			envs[key] = val
		}
	}
	return envs
}

// getPort extracts the port from a host:port string
func getPort(hostPort string) int {
	_, _port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(_port)
	if err != nil {
		return 0
	}
	return port
}
