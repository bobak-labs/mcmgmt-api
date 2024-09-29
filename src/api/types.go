package api

import (
	"github.com/bobak-labs/mcmgmt-api/services/backup"
	"github.com/bobak-labs/mcmgmt-api/services/container"
	login "github.com/bobak-labs/mcmgmt-api/services/login"
)

type JSONResponse struct {
	ResponseContent any    `json:"response"`
	HTTPStatus      int    `json:"http_status"`
	Message         string `json:"message"`
}

func NewJSONResponse(status int, msg string, content any) JSONResponse {
	return JSONResponse{
		ResponseContent: content,
		HTTPStatus:      status,
		Message:         msg,
	}
}

type ServerConfig struct {
	ListenPort string
	LogsPath   string
}

type APIServer struct {
	ServerConfig
	containerService *container.ContainerService
	backupService    *backup.BackupService
	loginService     *login.LoginService
	jwtSecret        []byte
}

func NewAPIServer(lp string, logsPath string, loginSvc *login.LoginService, r *container.ContainerService, b *backup.BackupService, secret string) *APIServer {
	return &APIServer{
		ServerConfig: ServerConfig{
			ListenPort: lp,
			LogsPath:   logsPath,
		},
		loginService:     loginSvc,
		containerService: r,
		backupService:    b,
		jwtSecret:        []byte(secret),
	}
}
