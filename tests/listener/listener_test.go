package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFlow(t *testing.T) {
	ctx := context.Background()

	net, err := network.New(ctx)
	require.NoError(t, err)
	defer func() {
		if err := net.Remove(ctx); err != nil {
			t.Fatalf("failed to remove network: %v", err)
		}
	}()

	//verne container
	verneReq := testcontainers.ContainerRequest{
		Image: "vernemq/vernemq:latest",
		Env: map[string]string{
			"DOCKER_VERNEMQ_ACCEPT_EULA":     "yes",
			"DOCKER_VERNEMQ_ALLOW_ANONYMOUS": "on",
		},
		ExposedPorts: []string{"1883/tcp"},
		WaitingFor:   wait.ForListeningPort("1883/tcp"),
		Networks:     []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {"vernemq"},
		},
	}
	verne, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: verneReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer verne.Terminate(ctx)
	verneHost, err := verne.Host(ctx)
	require.NoError(t, err)
	vernePort, err := verne.MappedPort(ctx, "1883")
	require.NoError(t, err)
	brokerAddr := verneHost + ":" + vernePort.Port()

	//gcp pubsub emulator
	pubsubReq := testcontainers.ContainerRequest{
		Image:        "gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators",
		Cmd:          []string{"gcloud", "beta", "emulators", "pubsub", "start", "--host-port=0.0.0.0:8085"},
		ExposedPorts: []string{"8085/tcp"},
		WaitingFor:   wait.ForListeningPort("8085/tcp"),
		Networks:     []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {"gcp-emulators"},
		},
	}
	pubsub, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: pubsubReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer pubsub.Terminate(ctx)
	pubHost, err := pubsub.Host(ctx)
	require.NoError(t, err)
	pubPort, err := pubsub.MappedPort(ctx, "8085")
	require.NoError(t, err)
	pubAddr := pubHost + ":" + pubPort.Port()

	//verne-listener container
	listenerReq := testcontainers.ContainerRequest{
		Image: "verne-listener:latest",
		Env: map[string]string{
			"MQTT_CONNECT_SUCCESS_PATH":   "/tmp/listener.status",
			"PUBSUB_CONNECT_SUCCESS_PATH": "/tmp/pub.status",
			"MQTT_TOPIC":                  "mqtt-source",
			"MQTT_BROKER_ADDRESS":         brokerAddr,
			"PUBSUB_HOST":                 pubAddr,
		},

		Files: []testcontainers.ContainerFile{
			{
				Reader:            bytes.NewReader([]byte("gcp-emulators")),
				ContainerFilePath: "/envs/host_ip",
				FileMode:          0644,
			},
		},
		Networks: []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {"listener"},
		},
	}
	listener, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: listenerReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer listener.Terminate(ctx)

	//log connection
	logs, _ := listener.Logs(ctx)
	buf := new(bytes.Buffer)
	buf.ReadFrom(logs)
	t.Log(buf.String())
}
