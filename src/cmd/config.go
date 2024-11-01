package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/common-nighthawk/go-figure"
)

type BindPath struct {
	Path  string
	Label string
}

type clientConfig struct {
	listenPort          string
	backupPath          string
	serverFilesPath     string
	trustedPath         string
	executorType        string
	memory              int64
	backupBindPath      BindPath
	serverFilesBindPath BindPath
}

func NewClientConfig(
	listenPort string,
	backupPath string,
	serverFilesPath string,
	trustedPath string,
	executorType string,
	memory int64) (*clientConfig, error) {

	backupBindPath := BindPath{
		Path:  filepath.Clean(backupPath),
		Label: "backup",
	}
	serverFilesBindPath := BindPath{
		Path:  filepath.Clean(serverFilesPath),
		Label: "serverfiles",
	}

	backupPath, err := verifyPath(trustedPath, backupBindPath)
	if err != nil {
		log.Fatalln("invalid backup path: ", backupBindPath, "error: ", err)
		return nil, err
	}

	serverFilesPath, err = verifyPath(trustedPath, serverFilesBindPath)
	if err != nil {
		log.Fatalln("invalid bind path: ", serverFilesBindPath, "error: ", err)
		return nil, err
	}

	if serverFilesPath == backupBindPath.Path {
		log.Fatalln("backup and server files paths cannot be the same")
		return nil, err
	}

	return &clientConfig{
		listenPort:          listenPort,
		backupPath:          backupPath,
		serverFilesPath:     serverFilesPath,
		trustedPath:         trustedPath,
		executorType:        executorType,
		memory:              memory,
		backupBindPath:      backupBindPath,
		serverFilesBindPath: serverFilesBindPath,
	}, nil
}

func NewDefaultClientConfig() (*clientConfig, error) {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalln("cannot fetch current directory: ", err)
		return nil, err
	}
	trustedPath, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("cannot fetch user home directory: ", err)
		return nil, err
	}

	defaultListenPort := ":7777"
	workdir := filepath.Dir(execPath)                           // working directory  (.)
	defaultBackupPath := fmt.Sprintf("%s/backups", workdir)     // folder in working directory (./backups)
	defaultServerFilesPath := fmt.Sprintf("%s/server", workdir) // working directory (./server)
	defaultMemory := 4                                          // in gigabytes
	defaultExecutorType := "docker"
	defaultBackupBindPath := BindPath{
		Path:  filepath.Clean(workdir),
		Label: "backup",
	}
	defaultServerFilesBindPath := BindPath{
		Path:  filepath.Clean(workdir),
		Label: "serverfiles",
	}

	return &clientConfig{
		listenPort:          defaultListenPort,
		backupPath:          defaultBackupPath,
		serverFilesPath:     defaultServerFilesPath,
		trustedPath:         trustedPath,
		executorType:        defaultExecutorType,
		memory:              int64(defaultMemory),
		backupBindPath:      defaultBackupBindPath,
		serverFilesBindPath: defaultServerFilesBindPath,
	}, nil
}

func (c *clientConfig) parseFlags() (*clientConfig, error) {
	fmt.Println("================================================================================================================")
	// fig1 := figure.NewFigure("welcome to:", "larry3d", true)
	// fig1.Print()
	fig2 := figure.NewFigure("mcmgmt-api", "larry3d", true)
	fig2.Print()
	fmt.Println("================================================================================================================")
	flag.Usage = c.flagUsage

	listenport := flag.Int("lp", 7777, "api server listen port")
	backuppath := flag.String("backups", c.backupPath, "Path where server backups are stored, any parent directory specified must exist.\n\tExample: If you want to store backups in /etc/minecraft/backups, directory /etc/minecraft MUST EXIST.")
	serverfiles := flag.String("server-files", c.serverFilesPath, "Path where server files are stored. Any parent directory specified must exist.\n\tExample: If you want to store backups in /etc/minecraft/serverfiles, directory /etc/minecraft MUST EXIST.")
	trustedpath := flag.String("trusted-path", c.trustedPath, "Boundary path to which user can traverse specifying backup/server directories.")
	executor := flag.String("executor", c.executorType, "Executor that will be used by the program to manage the server (native/docker/kubernetes)")
	memory := flag.Int64("memory", c.memory, "Memory for the minecraft server in gigabytes (GB).")

	flag.Parse()

	log.Printf("Started mcmgmt-api binary with PID %v\n", os.Getpid())
	log.Println("default backup path:", c.backupPath)
	log.Println("default server files path:", c.serverFilesPath)
	log.Println("default memory in gigabytes:", c.memory)
	log.Println("default trusted path:", c.trustedPath)
	log.Println("default executor:", c.executorType)

	flag.Visit(func(f *flag.Flag) {
		log.Printf("flag: %s overriden to: %s\n", f.Name, f.Value.String())
	})

	listenPortString := fmt.Sprintf(":%s", strconv.Itoa(*listenport))

	clientConf, err := NewClientConfig(listenPortString, *backuppath, *serverfiles, *trustedpath, *executor, *memory)
	if err != nil {
		return nil, err
	}

	return clientConf, nil
}

func (c *clientConfig) flagUsage() {
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
