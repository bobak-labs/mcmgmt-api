package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func InitBucket(bucketName, projectID, localBackupPath string) (*Bucket, error) {
	bucket, err := NewBucket(bucketName, projectID, localBackupPath)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	return bucket, nil
}

func InitRunner(containerImage, containerName, serverFilesPath string) *ContainerService {
	img := containerImage
	cn := containerName

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
		User:         fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()), // Match the current user
		Cmd:          strslice.StrSlice{"sleep", "3600000"},          // Keep container running
	}
	hostconf := container.HostConfig{
		Resources: container.Resources{
			Memory: 4 * 2147483648,
		},
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

	runner := NewContainerRunner(
		img,
		cn,
		networkName,
		conf,
		hostconf,
		netconf,
		platform,
		pullopts,
		startopts)

	return runner
}
