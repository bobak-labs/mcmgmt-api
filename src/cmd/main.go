package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	api "github.com/bobak-labs/mcmgmt-api/api"
	util "github.com/bobak-labs/mcmgmt-api/lib"
	backups "github.com/bobak-labs/mcmgmt-api/services/backup"
	login "github.com/bobak-labs/mcmgmt-api/services/login"
	"github.com/common-nighthawk/go-figure"
)

type BindPath struct {
	Path  string
	Label string
}

func main() {

	fmt.Println("================================================================================================================")
	// fig1 := figure.NewFigure("welcome to:", "larry3d", true)
	// fig1.Print()
	fig2 := figure.NewFigure("mcmgmt-api", "larry3d", true)
	fig2.Print()
	fmt.Println("================================================================================================================")
	flag.Usage = func() {
		var flags []flag.Flag
		var flagsStr string

		// Iterate over all flags and append them to the slice
		flag.VisitAll(func(f *flag.Flag) {
			flags = append(flags, *f) // De-reference and add flag to slice
		})

		for _, f := range flags {
			flagsStr += fmt.Sprintf("--%s, ", f.Name)
		}

		fmt.Fprintf(os.Stderr, "mcmgmt-api - web application for managing minecraft server.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "Available options: [%v]\n\n", flagsStr[:len(flagsStr)-2])
		fmt.Fprintf(os.Stderr, "Options:\n")
		// Print each flag in a custom format
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  --%s/-%s (default value=%v): \n\t%s\n\n\n", f.Name, f.Name, f.DefValue, f.Usage)
		})
		fmt.Fprintf(os.Stderr, "  --help/-help: \n\tShows help menu\n\n\n")
		os.Exit(0) // exit after showing help
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalln("cannot fetch current directory: ", err)
	}
	trustedPath, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("cannot fetch user home directory: ", err)
	}

	workdir := filepath.Dir(execPath)                       // working directory  (.)
	defaultBackupPath := fmt.Sprintf("%s/backups", workdir) // folder in working directory (./backups)
	defaultServerFilesPath := workdir                       // working directory (.)
	defaultMemory := 4                                      // in gigabytes

	listenport := flag.Int("lp", 7777, "api server listen port")
	backuppath := flag.String("backups", defaultBackupPath, "Path where server backups are stored, any parent directory specified must exist.\n\tExample: If you want to store backups in /etc/minecraft/backups, directory /etc/minecraft MUST EXIST.")
	serverfiles := flag.String("server-files", defaultServerFilesPath, "Path where server files are stored. Any parent directory specified must exist.\n\tExample: If you want to store backups in /etc/minecraft/serverfiles, directory /etc/minecraft MUST EXIST.")
	trustedpath := flag.String("trusted-path", trustedPath, "Boundary path to which user can traverse specifying backup/server directories.")
	memory := flag.Int64("memory", int64(defaultMemory), "Memory for the minecraft server in gigabytes (GB).")

	flag.Parse()

	log.Printf("Started mcmgmt-api binary with PID %v\n", os.Getpid())
	log.Println("default backup path:", defaultBackupPath)
	log.Println("default server files path:", defaultServerFilesPath)
	log.Println("default memory in gigabytes:", defaultMemory)
	log.Println("default trusted path:", trustedPath)

	// server params
	listenPort := fmt.Sprintf(":%d", *listenport)
	backupBindPath := BindPath{
		Path:  filepath.Clean(*backuppath),
		Label: "backup",
	}
	serverFilesBindPath := BindPath{
		Path:  filepath.Clean(*serverfiles),
		Label: "serverfiles",
	}

	trustedPath = *trustedpath

	flag.Visit(func(f *flag.Flag) {
		log.Printf("flag: %s overriden to: %s\n", f.Name, f.Value.String())
	})

	// fmt.Println(listenPort, backupBindPath, serverFilesBindPath)
	backupPath, err := verifyPath(trustedPath, backupBindPath)
	if err != nil {
		log.Fatalln("invalid backup path: ", backupBindPath, "error: ", err)
	}

	serverFilesPath, err := verifyPath(trustedPath, serverFilesBindPath)
	if err != nil {
		log.Fatalln("invalid bind path: ", serverFilesBindPath, "error: ", err)
	}

	if serverFilesPath == backupBindPath.Path {
		log.Fatalln("backup and server files paths cannot be the same")
	}

	// container params
	img := "itzg/minecraft-server"
	cn := fmt.Sprintf("mcserver-%s", util.RandomString(9))

	// needed envs
	secret := os.Getenv("JWT_SECRET")
	bucketName := os.Getenv("BACKUPS_BUCKET")
	projectID := os.Getenv("PROJECT_ID")

	//init server

	logPath := fmt.Sprintf("%v/mcdata/logs/latest.log", serverFilesPath)

	// create login service
	loginSvc := login.NewLoginService(secret)

	// create runner
	runner := InitRunner(img, cn, serverFilesPath, *memory)

	// create bucket controller
	bucket, err := InitBucket(bucketName, projectID, backupPath)
	if err != nil {
		log.Fatalln(err)
	}

	backupSvc := backups.NewBackupService(bucket, backupPath)

	// create API server instance
	server := api.NewAPIServer(listenPort, logPath, loginSvc, runner, backupSvc, secret)
	server.Run()
}
