package main

import (
	"fmt"
	"log"
	"os"

	api "github.com/bobak-labs/mcmgmt-api/api"
	util "github.com/bobak-labs/mcmgmt-api/lib"
	backups "github.com/bobak-labs/mcmgmt-api/services/backup"
	login "github.com/bobak-labs/mcmgmt-api/services/login"
)

func main() {

	// container params
	img := "itzg/minecraft-server"
	cn := fmt.Sprintf("mcserver-%s", util.RandomString(9))

	// needed envs
	secret := os.Getenv("JWT_SECRET")
	bucketName := os.Getenv("BACKUPS_BUCKET")
	projectID := os.Getenv("PROJECT_ID")

	// create default configuration
	defaultConfig, err := NewDefaultClientConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// override configuration parsing input flags
	customConfig, err := defaultConfig.parseFlags()
	if err != nil {
		log.Fatalln(err)
	}

	//init server
	logPath := fmt.Sprintf("%v/mcdata/logs/latest.log", customConfig.serverFilesPath)

	// create login service
	loginSvc := login.NewLoginService(secret)

	// create executor (native, docker, k8s)
	runner := InitExecutorService(customConfig.executorType, img, cn, customConfig.serverFilesPath, customConfig.memory)

	// create bucket controller
	bucket, err := InitBucket(bucketName, projectID, customConfig.backupPath)
	if err != nil {
		log.Fatalln(err)
	}

	backupSvc := backups.NewBackupService(bucket, customConfig.backupPath)

	// create API server instance
	server := api.NewAPIServer(customConfig.listenPort, logPath, loginSvc, runner, backupSvc, secret)
	server.Run()
}
