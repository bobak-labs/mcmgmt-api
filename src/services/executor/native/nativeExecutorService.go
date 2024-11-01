package native

import (
	"log"

	"github.com/bobak-labs/mcmgmt-api/services/executor"
)

type NativeExecutorService struct {
}

type NativeResourceData struct {
}

func (nex *NativeExecutorService) StartServer() (*executor.ExecutionResponse, error) {
	log.Println("unimplemented")
	return nil, nil
}

func (nex *NativeExecutorService) StopServer() (*executor.ExecutionResponse, error) {
	log.Println("unimplemented")
	return nil, nil
}

func (nex *NativeExecutorService) GetStatus() (*executor.ExecutionResponse, error) {
	log.Println("unimplemented")
	return nil, nil
}
