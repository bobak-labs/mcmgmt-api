package main

import (
	"fmt"
	"log"
	"math/rand/v2"

	backups "github.com/bobak-labs/mcmgmt-api/services/backup"
	"github.com/bobak-labs/mcmgmt-api/services/executor"
	dockerexec "github.com/bobak-labs/mcmgmt-api/services/executor/docker"
	k8sexec "github.com/bobak-labs/mcmgmt-api/services/executor/kubernetes"
	nativeexec "github.com/bobak-labs/mcmgmt-api/services/executor/native"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func InitBucket(bucketName, projectID, localBackupPath string) (*backups.Bucket, error) {
	bucket, err := backups.NewBucket(bucketName, projectID, localBackupPath)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	return bucket, nil
}

func InitExecutorService(executorType, containerImage, containerName, serverFilesPath string, memory int64) executor.ExecutorService {
	totalMemory := memory * 1024 * 1024 * 1024

	var executor executor.ExecutorService
	switch executorType {
	case "native":
		executor = InitNativeExecutor()
	case "docker":
		executor = InitDockerExecutor(containerImage, containerName, serverFilesPath, totalMemory)
	case "kubernetes":
		executor = InitKubernetesExecutor()
	}

	return executor
}

func InitNativeExecutor() executor.ExecutorService {
	return &nativeexec.NativeExecutorService{}
}

func InitDockerExecutor(img, cn, serverFilesPath string, totalMemory int64) executor.ExecutorService {
	ports, err := nat.NewPort("tcp", "25565-25565")
	if err != nil {
		panic(err)
	}
	networkName := fmt.Sprintf("mcnet-%d", rand.IntN(10000))

	conf := container.Config{
		Hostname:     "minecraft",
		Image:        img,
		ExposedPorts: nat.PortSet{ports: struct{}{}},
		Env:          []string{"EULA=TRUE"},
		// User:         fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), // Match the current user
		// Cmd: strslice.StrSlice{"sleep", "3600000"}, // Keep container running - for tests
	}
	hostconf := container.HostConfig{
		Resources: container.Resources{
			Memory: totalMemory,
		},
		//todo: create a way to also support docker volumes
		Binds: []string{
			fmt.Sprintf("%v/mcdata:/data", serverFilesPath),
		},
		PortBindings: nat.PortMap{
			"25565/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "25565",
				},
			},
		},
		AutoRemove:  false,
		NetworkMode: container.NetworkMode(container.NetworkMode(networkName).NetworkName()),
		Privileged:  true,
	}

	netconf := network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {
				NetworkID: networkName,
			},
		},
	}
	platform := v1.Platform{}
	pullopts := types.ImagePullOptions{}
	startopts := types.ContainerStartOptions{}

	if !hostconf.AutoRemove {
		hostconf.AutoRemove = true
	}

	dockerExecutor := dockerexec.NewDockerExecutorService(
		img,
		cn,
		networkName,
		conf,
		hostconf,
		netconf,
		platform,
		pullopts,
		startopts)

	return dockerExecutor
}

func InitKubernetesExecutor() executor.ExecutorService {
	return &k8sexec.KubernetesExecutorService{}
}
