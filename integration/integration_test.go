//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var registryHost string

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start registry container
	registry, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "registry:2",
			ExposedPorts: []string{"5000/tcp"},
			WaitingFor:   wait.ForHTTP("/v2/").WithPort("5000/tcp").WithStartupTimeout(30 * time.Second),
			Env: map[string]string{
				"REGISTRY_STORAGE_DELETE_ENABLED": "true",
			},
		},
		Started: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start registry: %v\n", err)
		os.Exit(1)
	}

	host, err := registry.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get registry host: %v\n", err)
		registry.Terminate(ctx)
		os.Exit(1)
	}

	port, err := registry.MappedPort(ctx, "5000")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get registry port: %v\n", err)
		registry.Terminate(ctx)
		os.Exit(1)
	}

	registryHost = fmt.Sprintf("%s:%s", host, port.Port())

	// Run testscript
	exitCode := testscript.RunMain(m, map[string]func() int{
		"blob": run,
	})

	registry.Terminate(ctx)
	os.Exit(exitCode)
}

func TestCLI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
		Setup: func(env *testscript.Env) error {
			env.Setenv("REGISTRY", registryHost)
			env.Setenv("XDG_CONFIG_HOME", env.WorkDir+"/.config")
			env.Setenv("XDG_CACHE_HOME", env.WorkDir+"/.cache")
			return copyTestData(env.WorkDir)
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"gentag": cmdGenTag,
		},
	})
}
