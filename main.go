package main

import (
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
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

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		type parameters struct {
			Body string `json:"body"`
			UserId string `json:"user_id"`
		} 

		type chirp struct {
			ID uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			UserId uuid.UUID `json:"user_id"`
			Body string `json:"body"`
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

		body := CleanString(params.Body)
		user_id, err := uuid.Parse(params.UserId)

		if err != nil {
			JsonError(w, err)	
			return
		}

		response, err := dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{Body: body, UserID: user_id})
		
		if err != nil {
			JsonError(w, err)
			return
		}


		chirp_json := chirp {
			Body: response.Body,
			ID: response.ID,
			UserId: response.UserID,
			CreatedAt: response.CreatedAt,
			UpdatedAt: response.UpdatedAt,
		}

		response_json , err := json.Marshal(chirp_json)

		if err != nil {
			JsonError(w, err)
			return
		}


		w.WriteHeader(201)
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

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type","application/json")

		type chirp struct {
			ID uuid.UUID `json:"id"`
			Body string `json:"body"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			UserID uuid.UUID `json:"user_id"`
		}

		chirps, err := dbQueries.GetAllChirps(req.Context())

		if err != nil {
			JsonError(w, err)
			return
		}

		response_chirps := []chirp {}

		for _, user_chirp := range chirps {
			response_chirps = append(response_chirps, chirp{
				ID: user_chirp.ID,
				Body: user_chirp.Body,
				CreatedAt: user_chirp.CreatedAt,
				UpdatedAt: user_chirp.UpdatedAt,
				UserID: user_chirp.UserID,
			})
		}

		chirp_response, err := json.Marshal(response_chirps)

		if err != nil {
			JsonError(w, err)
			return
		}

		w.WriteHeader(200)
		w.Write(chirp_response)
	})

	server.ListenAndServe()
}
