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
		platform := os.Getenv("PLATFORM")
		
		if platform != "dev" {
			w.WriteHeader(403)
			w.Write([]byte("Deletion failed. Perhaps you tried to delete from outside local host?\n"))
		} else {
			w.WriteHeader(200)
			dbQueries.RemoveAllUsers(req.Context())
			w.Write([]byte("All users removed from users DB."))
		}
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

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
			JsonError(w, err)
			return 
		}

		if len(params.Body) > 140 {
			w.WriteHeader(400)		
			w.Write([]byte(`{"error": "Chirp is too long"}`))
			return
		}

		response := CleanString(params.Body)
		response_json, _ := json.Marshal(cleaned {Cleaned_body: response})
		
		w.WriteHeader(200)
		w.Write(response_json)
	})


	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		type parameters struct {
			Email string `json:"email"`
		} 

		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)

		if err != nil {
			JsonError(w, err)
			return
		}

		chirpy_user, err := dbQueries.CreateUser(req.Context(), params.Email)

		if err != nil {
			JsonError(w, err)
			return
		}

		userResponse := User{
			ID: chirpy_user.ID,
			CreatedAt: chirpy_user.CreatedAt,
			UpdatedAt: chirpy_user.UpdatedAt,
			Email: chirpy_user.Email,
		}

		data, err := json.Marshal(userResponse)

		if err != nil {
			JsonError(w, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(data)
	})

	server.ListenAndServe()
}
