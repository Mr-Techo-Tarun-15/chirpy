package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"

	"github.com/Mr-Techo-Tarun-15/chirpy/healthz"
	"github.com/Mr-Techo-Tarun-15/chirpy/internal/database"
)

var newHTTPServeMux *http.ServeMux = http.NewServeMux()
var newHTTPServer http.Server = http.Server{
	Handler: newHTTPServeMux,
	Addr:    ":8080",
}

func errorCase(w http.ResponseWriter, errorStatusCode int, processError, systemError string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorStatusCode)
	json.NewEncoder(w).Encode(map[string]string{"process error": processError, "system error": systemError})
}

type apiConfig struct {
	fileserverHits  atomic.Int32
	databaseQueries *database.Queries
	currentPlatform string
	secretForJWT    string
	polkaKey        string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	crtPltfrm := os.Getenv("PLATFORM")
	JWTSecret := os.Getenv("SECRET")
	polkakey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	healthz.ReadinessEndpoint(newHTTPServeMux)
	filePath := http.FileServer(http.Dir("."))

	cfg := &apiConfig{
		databaseQueries: dbQueries,
		currentPlatform: crtPltfrm,
		secretForJWT:    JWTSecret,
		polkaKey:        polkakey,
	}

	// Get Handlers
	newHTTPServeMux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	newHTTPServeMux.HandleFunc("GET /api/chirps", cfg.getChirpsHandler)
	newHTTPServeMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpFromIDHandler)

	// Put Handlers
	newHTTPServeMux.HandleFunc("PUT /api/users", cfg.updateUsersHandler)

	// Post Handlers
	newHTTPServeMux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	newHTTPServeMux.HandleFunc("POST /api/users", cfg.createUsersHandler)
	newHTTPServeMux.HandleFunc("POST /api/chirps", cfg.postChirpsHandler)
	newHTTPServeMux.HandleFunc("POST /api/login", cfg.loginHandler)
	newHTTPServeMux.HandleFunc("POST /api/refresh", cfg.refreshHandler)
	newHTTPServeMux.HandleFunc("POST /api/revoke", cfg.revokeHandler)
	newHTTPServeMux.HandleFunc("POST /api/polka/webhooks", cfg.webhooksHandler)

	// Delete Handlers
	newHTTPServeMux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpFromIDHandler)

	// Actual App Handlers
	newHTTPServeMux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", filePath)))

	err = newHTTPServer.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
