package executor

type ExecutionResponse struct {
	Response any
}

type ExecutorService interface {
	StartServer() (*ExecutionResponse, error)
	StopServer() (*ExecutionResponse, error)
	GetStatus() (*ExecutionResponse, error)
}
