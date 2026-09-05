package main

import (
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL) 

	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(db)
	apiCfg := &apiConfig{}
	apiCfg.db = dbQueries
	mux := http.NewServeMux()
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	
	mux.Handle("GET /app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(
			`
			<html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>
			`, 
			apiCfg.returnRequests())))
	}) 
	
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		apiCfg.resetRequests()
		w.Write([]byte(fmt.Sprintf("Server Hits reset! %d", apiCfg.returnRequests())))
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, req *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		} 

		type cleaned struct {
			Cleaned_body string `json:"cleaned_body"`
		}

		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)

		if err != nil {
			error_response := `{"error": "Something went wrong"}`
			fmt.Printf("Error decoding parameters: %s", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte(error_response))
			return 
		}

		if len(params.Body) > 140 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)		
			w.Write([]byte(`{"error": "Chirp is too long"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := CleanString(params.Body)
		response_json, _ := json.Marshal(cleaned {Cleaned_body: response})
		
		w.WriteHeader(200)
		w.Write(response_json)
	})

	server.ListenAndServe()
}
