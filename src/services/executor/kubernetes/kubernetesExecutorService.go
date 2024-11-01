package kubernetes

import (
	"log"

	"github.com/bobak-labs/mcmgmt-api/services/executor"
)

type KubernetesExecutorService struct {
}

type KubernetesResourceData struct {
}

func (kex *KubernetesExecutorService) StartServer() (*executor.ExecutionResponse, error) {
	log.Println("unimplemented")
	return nil, nil
}

func (kex *KubernetesExecutorService) StopServer() (*executor.ExecutionResponse, error) {
	log.Println("unimplemented")
	return nil, nil
}

func (kex *KubernetesExecutorService) GetStatus() (*executor.ExecutionResponse, error) {
	log.Println("unimplemented")
	return nil, nil
}
