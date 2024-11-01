package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/bobak-labs/mcmgmt-api/docs"
	util "github.com/bobak-labs/mcmgmt-api/lib"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (s *APIServer) Run() {

	r := mux.NewRouter()

	r.Use(corsMiddleware, LoggerMiddleware)
	// Serve Swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.HandleFunc("/login", s.LoginHandler).Methods("POST")

	r.Handle("/stop", s.JwtAuth(http.HandlerFunc(s.StopHandler))).Methods("POST")
	r.Handle("/start", s.JwtAuth(http.HandlerFunc(s.StartHandler))).Methods("POST")

	r.Handle("/logs", s.JwtAuth(http.HandlerFunc(s.LogsHandler))).Methods("GET")

	r.Handle("/backup", s.JwtAuth(http.HandlerFunc(s.BackupHandler))).Methods("POST")
	r.Handle("/backup", s.JwtAuth(http.HandlerFunc(s.GetBackupHandler))).Methods("GET")

	r.Handle("/backup/delete", s.JwtAuth(http.HandlerFunc(s.DeleteBackupHandler))).Methods("DELETE")
	r.Handle("/backup/load", s.JwtAuth(http.HandlerFunc(s.LoadBackupHandler))).Methods("POST")

	r.Handle("/sync", s.JwtAuth(http.HandlerFunc(s.SyncHandler))).Methods("POST")

	log.Printf("Server listening on port %v\n", s.ListenPort)
	if err := http.ListenAndServe(s.ListenPort, r); err != nil {
		panic(err)
	}
}

// LoadBackupHandler loads a backup from a file or multipart form data
// @Summary Load a backup
// @Description Load a backup from the disk or multipart form data
// @Tags backup
// @Accept  json
// @Produce  json
// @Param file query string false "Whether to load backup from a file"
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /backup/load [post]
func (s *APIServer) LoadBackupHandler(w http.ResponseWriter, r *http.Request) {
	// To load backup:
	// shutdown server
	// format the time to match the desired format
	// backup current state
	// remove current server files
	// unzip backup to mcdata/
	// start the server
	var returnData any
	doesExist, _ := util.Exists(s.backupService.BackupPath)
	if !doesExist {
		if err := os.Mkdir(s.backupService.BackupPath, os.FileMode(0755)); err != nil {
			log.Println("cannot create directory", err)
			WriteJSON(w, MessageToJSON(http.StatusInternalServerError, err.Error(), nil))
			return
		}
	}

	// Stop the server container
	if _, err := s.executorService.StopServer(); err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, err.Error(), nil))
		return
	}

	fileFlag := r.URL.Query().Get("file")

	if fileFlag == "true" {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<30) // 1 GB limit

		// Parse multipart form data
		if err := r.ParseMultipartForm(1 << 30); err != nil { // 1 GB limit
			WriteJSON(w, MessageToJSON(http.StatusBadRequest, "unable to parse multipart form", nil))
			return
		}

		// Get the file from the form
		file, header, err := r.FormFile("file")
		if err != nil {
			WriteJSON(w, MessageToJSON(http.StatusBadRequest, "error retrieving the file", nil))
			return
		}
		defer file.Close()

		// Extract the real file name
		fileName := header.Filename

		// Create a progress logger to log the upload status
		progressReader := &util.ProgressReader{
			Reader:      file,
			TotalBytes:  header.Size,
			LoggedBytes: 0,
			Logger:      log.Default(),
		}

		data, err := s.backupService.UploadBackupMultipart(progressReader, fileName)
		if err != nil {
			WriteJSON(w, MessageToJSON(http.StatusInternalServerError, err.Error(), nil))
			return
		}
		returnData = data

	} else {
		var backupFileName struct {
			Backup string `json:"backup"`
		}

		// Decode JSON body to get the backup file name
		if err := json.NewDecoder(r.Body).Decode(&backupFileName); err != nil {
			WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "error decoding body", nil))
			return
		}
		backupFile := backupFileName.Backup

		data, err := s.backupService.LoadBackupFromDisk(backupFile)
		if err != nil {
			WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "loading data from disk failed", nil))
			return
		}

		if _, err := s.executorService.StartServer(); err != nil {
			WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "failed to start server", nil))
			return
		}
		returnData = data
	}

	WriteJSON(w, MessageToJSON(http.StatusOK, "loading data successful", returnData))
}

// SyncHandler synchronizes data
// @Summary Sync data
// @Description Sync the latest data
// @Tags backup
// @Accept  json
// @Produce  json
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /sync [post]
func (s *APIServer) SyncHandler(w http.ResponseWriter, r *http.Request) {

	backupData, err := s.backupService.Sync()
	if err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "error syncing data", nil))
	}
	WriteJSON(w, MessageToJSON(http.StatusOK, "synced successfully", backupData))

}

// LoginHandler handles user login
// @Summary Login
// @Description Authenticates the user and returns a JWT token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param credentials body map[string]string true "User credentials"
// @Success 200 {object} JSONResponse
// @Failure 400 {object} JSONResponse
// @Failure 401 {object} JSONResponse
// @Router /login [post]
func (s *APIServer) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		WriteJSON(w, MessageToJSON(http.StatusBadRequest, "invalid request payload", nil))
		return
	}

	// Extract the username and password
	username := credentials.Username
	password := credentials.Password

	if username == "" || password == "" {
		WriteJSON(w, MessageToJSON(http.StatusUnauthorized, "login information incomplete", nil))
		return
	}

	tokenString, expirationTime, err := s.loginService.Login(username, password)
	if err != nil {
		WriteJSON(w, MessageToJSON(http.StatusUnauthorized, "unauthorized", nil))
		return
	}

	// Set the Authorization header in the response
	w.Header().Set("Authorization", "Bearer "+*tokenString)
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"token":          *tokenString,
		"expirationTime": expirationTime.Format(time.RFC3339),
	}
	// Return the token and expiration time in the response body
	WriteJSON(w, MessageToJSON(http.StatusOK, "authorized", response))
}

// GetBackupHandler retrieves the list of available backups
// @Summary Get backups
// @Description Retrieves the list of available backups
// @Tags backup
// @Accept  json
// @Produce  json
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /backup [get]
func (s *APIServer) GetBackupHandler(w http.ResponseWriter, r *http.Request) {
	backups, err := s.backupService.GetBackups()
	if err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "getting backups failed", nil))
		return
	}
	response := map[string][]string{
		"backups": backups,
	}
	WriteJSON(w, MessageToJSON(http.StatusOK, "retrieved backups", response))
}

// BackupHandler creates a backup
// @Summary Create a backup
// @Description Creates a backup of the server
// @Tags backup
// @Accept  json
// @Produce  json
// @Param backup body map[string]string true "Backup information"
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /backup [post]
func (s *APIServer) BackupHandler(w http.ResponseWriter, r *http.Request) {
	var backupFileName struct {
		Backup string `json:"backup"`
	}

	// Decode JSON body to get the backup file name
	if err := json.NewDecoder(r.Body).Decode(&backupFileName); err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, err.Error(), nil))
		return
	}
	backupName := backupFileName.Backup
	if backupName == "" {
		backupName = "server"
	}

	backupData, err := s.backupService.Backup(backupName)
	if err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, err.Error(), nil))
		return
	}
	WriteJSON(w, MessageToJSON(http.StatusOK, "backup successful", backupData))
}

// DeleteBackupHandler deletes a backup
// @Summary Delete a backup
// @Description Deletes a specified backup
// @Tags backup
// @Accept  json
// @Produce  json
// @Param delete query string true "Name of the backup to delete"
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /backup/delete [delete]
func (s *APIServer) DeleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	backupToDelete := r.URL.Query().Get("delete")

	backupData, err := s.backupService.DeleteLocalBackup(backupToDelete)
	if err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, err.Error(), nil))
		return
	}

	WriteJSON(w, MessageToJSON(http.StatusOK, fmt.Sprintf("backup %s deleted", backupToDelete), backupData))
}

// LogsHandler retrieves server logs
// @Summary Get logs
// @Description Retrieves the server logs
// @Tags logs
// @Accept  json
// @Produce  json
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /logs [get]
func (s *APIServer) LogsHandler(w http.ResponseWriter, r *http.Request) {
	logsPath := s.LogsPath
	logs, err := util.ReadLinesFromFile(logsPath)
	if err != nil {
		resp := map[string]string{
			"logs": "error reading logs",
		}
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "error reading logs", resp))
		return
	}

	resp := map[string][]string{
		"logs": logs,
	}

	WriteJSON(w, MessageToJSON(http.StatusOK, "logs retrieved", resp))

}

// StopHandler stops the server
// @Summary Stop server
// @Description Stops the server container
// @Tags server
// @Accept  json
// @Produce  json
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /stop [post]
func (s *APIServer) StopHandler(w http.ResponseWriter, r *http.Request) {

	data, err := s.executorService.StopServer()
	if err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "internal server error", nil))
		return
	}

	WriteJSON(w, MessageToJSON(http.StatusOK, "server stopped", data))
}

// StartHandler starts the server
// @Summary Start server
// @Description Starts the server container
// @Tags server
// @Accept  json
// @Produce  json
// @Success 200 {object} JSONResponse
// @Failure 500 {object} JSONResponse
// @Router /start [post]
func (s *APIServer) StartHandler(w http.ResponseWriter, r *http.Request) {

	data, err := s.executorService.StartServer()
	if err != nil {
		WriteJSON(w, MessageToJSON(http.StatusInternalServerError, "internal server error", data))
		return
	}

	WriteJSON(w, MessageToJSON(http.StatusOK, "server started", data))

}
