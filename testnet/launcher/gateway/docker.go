package gateway

import (
	"fmt"
	"time"

	"github.com/sanity-io/litter"
	"github.com/ten-protocol/go-ten/go/common/docker"
	"github.com/ten-protocol/go-ten/go/common/retry"
	"github.com/valyala/fasthttp"
)

type DockerGateway struct {
	cfg *Config
}

func NewDockerGateway(cfg *Config) (*DockerGateway, error) {
	return &DockerGateway{
		cfg: cfg,
	}, nil // todo (@pedro) - add validation
}

func (n *DockerGateway) Start() error {
	fmt.Printf("Starting gateway with config: \n%s\n\n", litter.Sdump(*n.cfg))

	cmds := []string{
		"ego",
		"run",
		"/home/ten/go-ten/tools/walletextension/main/main",
		fmt.Sprintf("-host=%s", "0.0.0.0"),
		fmt.Sprintf("-port=%d", n.cfg.gatewayHTTPPort),
		fmt.Sprintf("-portWS=%d", n.cfg.gatewayWSPort),
		fmt.Sprintf("-nodePortHTTP=%d", n.cfg.tenNodeHTTPPort),
		fmt.Sprintf("-nodePortWS=%d", n.cfg.tenNodeWSPort),
		fmt.Sprintf("-nodeHost=%s", n.cfg.tenNodeHost),
		"-dbType=sqlite",
		"-logPath=gateway_logs.log",
		fmt.Sprintf("-rateLimitUserComputeTime=%d", n.cfg.rateLimitUserComputeTime),
	}

	// Set environment variables as map[string]string
	envs := map[string]string{
		"OE_SIMULATION": "1", // Set to "0" if not in simulation mode
	}

	// Map required devices as map[string]string
	devices := map[string]string{}

	// No volume mappings
	volumes := map[string]string{}

	// Start the Docker container with updated settings
	_, err := docker.StartNewContainer(
		"gateway",
		n.cfg.dockerImage,
		cmds,
		[]int{n.cfg.gatewayHTTPPort, n.cfg.gatewayWSPort},
		envs,
		devices,
		volumes,
		true, // autoRestart
	)
	return err
}

func (n *DockerGateway) IsReady() error {
	timeout := time.Minute
	interval := time.Second

	return retry.Do(func() error {
		statusCode, _, err := fasthttp.Get(nil, fmt.Sprintf("http://127.0.0.1:%d/v1/health/", n.cfg.gatewayHTTPPort))
		if err != nil {
			return err
		}

		if statusCode != fasthttp.StatusOK {
			return fmt.Errorf("status not ok - status received: %s", fasthttp.StatusMessage(statusCode))
		}

		return nil
	}, retry.NewTimeoutStrategy(timeout, interval))
}
