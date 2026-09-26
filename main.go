package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
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
	apiCfg.jwt_secret = os.Getenv("JWT_SECRET")
	apiCfg.polka_api_key = os.Getenv("POLKA_KEY")
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
			IsChirpyRed bool `json:"is_chirpy_red"`
		} 


		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err = decoder.Decode(&params)

		if err != nil {
			JsonError(w, err)
			return 
		}

		bearer_token, err := auth.GetBearerToken(req.Header)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("No bearer token present probably: %v", err)))
		}

		user_id, err := auth.ValidateJWT(bearer_token, apiCfg.jwt_secret)


		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("User token could not be validated by validation function: %v", err)))
			return
		}



		if len(params.Body) > 140 {
			w.WriteHeader(400)		
			w.Write([]byte(`{"error": "Chirp is too long"}`))
			return
		}

		body := CleanString(params.Body)
		response, err := dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{Body: body, UserID: user_id})
		
		if err != nil {
			JsonError(w, err)
			return
		}


		chirp_json := chirp {
			Body: response.Body,
			ID: response.ID,
			UserID: response.UserID,
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
			Password string `json:"password"`
		} 

		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)

		if err != nil {
			JsonError(w, err)
			return
		}

		hashed_password, err := auth.HashPassword(params.Password)

		if err != nil {
			JsonError(w, err)
			return
		}
		
		chirpy_user, err := dbQueries.CreateUser(req.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashed_password})

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

	mux.HandleFunc("GET /api/chirps/{chirpID}", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-type", "application/json")
		chirpID, err := uuid.Parse(req.PathValue("chirpID"))

		if err != nil {
			JsonError(w, err)
			return
		}

		wanted_chirp, err := dbQueries.GetChirpByID(req.Context(), chirpID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				error_response := `{"error": "Chirp Not Found!"}`
				fmt.Printf("Chirp not found in database!\n ErrorCode: %s", err)
				w.WriteHeader(404)
				w.Write([]byte(error_response))
				return
			}

			JsonError(w, err)
			return
		}

		chirp_response, err := json.Marshal(chirp {
			ID: wanted_chirp.ID,
			Body: wanted_chirp.Body,
			CreatedAt: wanted_chirp.CreatedAt,
			UpdatedAt: wanted_chirp.UpdatedAt,
			UserID: wanted_chirp.UserID,
		})

		if err != nil {
			JsonError(w, err)
			return
		}

		w.WriteHeader(200)
		w.Write(chirp_response)
	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, req *http.Request){
		type parameters struct {
			Password string `json:"password"`
			Email string `json:"email"`
			IsChirpyRed bool `json:"is_chipry_red"`
		}

		params := parameters{}
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&params)

		if err != nil {
			JsonError(w, err)
			return
		}

		user_data, err := dbQueries.GetUserByEmail(req.Context(), params.Email)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Incorrect email or password"))
			return
		}

		hashed_password, err := auth.CheckPasswordHash(params.Password, user_data.HashedPassword)

		if err != nil {
			JsonError(w, err)
			return
		}

		if hashed_password != true {
			w.WriteHeader(401)
			w.Write([]byte("Incorrect email or password"))
			return
		}
		
		token, err := auth.MakeJWT(user_data.ID, apiCfg.jwt_secret, time.Hour)

		if err != nil {
			JsonError(w, err)
			return
		}

		refresh_token, err := auth.MakeRefreshToken()

		if err != nil {
			JsonError(w, err)
			return
		}

		_, err = dbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{Token: refresh_token, UserID: user_data.ID})

		if err != nil {
			JsonError(w, err)
			return
		}

		user_response, err := json.Marshal(TokenUser{
			ID: user_data.ID,
			CreatedAt: user_data.CreatedAt,
			UpdatedAt: user_data.UpdatedAt,
			Email: user_data.Email,
			Token: token,
			RefreshToken: refresh_token,
			IsChirpyRed: user_data.IsChirpyRed,
		})

		if err != nil {
			JsonError(w, err)
			return
		}

		w.WriteHeader(200)
		w.Write(user_response)
	})

	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, req *http.Request) {
		bearerToken, err := auth.GetBearerToken(req.Header)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Refresh route failed to extract refresh token: %v", err)))
			return
		}

		userId, err := dbQueries.UserFromRefreshToken(req.Context(), bearerToken)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Invalid Refresh Token: %v", err)))
			return
		}

		access_token, err := auth.MakeJWT(userId, apiCfg.jwt_secret, time.Hour)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Access Token Generation Failed At Refresh Token Route!: %v", err)))
			return
		}

		access_token_marshal, err := json.Marshal(struct{Token string `json:"token"`} {Token: access_token})

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Access Token Marshalling Failed At RefreshToken Route!: %v", err)))
			return
		}

		w.WriteHeader(200)
		w.Write(access_token_marshal)
	})

	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, req *http.Request) {
		bearerToken, err := auth.GetBearerToken(req.Header)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("No refresh Token Present In Refresh Token Revoke Route: %v", err)))
			return
		}

		err = dbQueries.RevokeRefreshToken(req.Context(), bearerToken)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Refresh Token Revocation Failed: %v", err)))
			return
		}

		w.WriteHeader(204)
	})

	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, req *http.Request) {
		bearerToken, err := auth.GetBearerToken(req.Header)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Access Token Not Present In PUT /api/users Route: %v", err)))
			return
		}

		type parameters struct {
			Password string `json:"password"`
			Email string `json:"email"`
			IsChirpyRed bool `json:"is_chipry_red"`
		}

		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err = decoder.Decode(&params)

		if err != nil {
			JsonError(w, err)
			return 
		}

		userID, err := auth.ValidateJWT(bearerToken, apiCfg.jwt_secret)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Invalid Auth Token: %v", err)))
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)

		if err != nil {
			JsonError(w, err)
			return
		}

		err = dbQueries.UpdateUserEmailAndPassword(req.Context(), database.UpdateUserEmailAndPasswordParams{ID: userID, Email: params.Email, HashedPassword: hashedPassword})

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Updating User Email And Password Failed: %v", err)))
		}

		userParams, err := json.Marshal(params)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Marshalling params for return object failed: %v", err)))
		}


		w.WriteHeader(200)
		w.Write(userParams)
	})

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", func(w http.ResponseWriter, req *http.Request) {
		bearerToken, err := auth.GetBearerToken(req.Header)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Bearer Token Not Present at DELETE /api/chirps: %v",err)))
			return
		}

		userID, err := auth.ValidateJWT(bearerToken, apiCfg.jwt_secret)

		if err != nil {
			w.WriteHeader(403)
			w.Write([]byte(fmt.Sprintf("Bearer Token Not Valid at DELETE /api/chirps: $v", err)))
			return
		}

		chirpID := req.PathValue("chirpID")
		parsedChirpID, err := uuid.Parse(chirpID);

		if err != nil {
			w.WriteHeader(403)
			w.Write([]byte(fmt.Sprintf("chirpId failed to parse to UUID datatype. Maybe checkfor incorrect or weird characters: %v", err)))
			return
		}

		chirp, err := dbQueries.GetChirpByID(req.Context(), parsedChirpID)

		if err != nil {
			w.WriteHeader(403)
			w.Write([]byte(fmt.Sprintf("Chirp Could Not Be Found By ID")))
			return
		}

		if chirp.UserID != userID {
			w.WriteHeader(403)
			w.Write([]byte(fmt.Sprintf("Chirp Not Made By User: %v")))
			return
		}


		err = dbQueries.DeleteChirp(req.Context(), database.DeleteChirpParams{ID: parsedChirpID, UserID: userID})

		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte(fmt.Sprintf("Chirp Deletion Failed. Check Chirp ID for accuracy: %v", err)))
			return
		}

		w.WriteHeader(204)
	})

	mux.HandleFunc("POST /api/polka/webhooks", func(w http.ResponseWriter, req *http.Request) {
		type userID struct {
			UserID string `json:"user_id"`
		}

		type parameters struct {
			Event string `json:"event"`
			Data userID `json:"data"`
		}

		decoder := json.NewDecoder(req.Body)
		params := parameters {}
		err = decoder.Decode(&params)

		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte(fmt.Sprintf("Input parameters is not in correct shape: %v", err)))
			return 
		}

		if params.Event != "user.upgraded" {
			w.WriteHeader(204)
			w.Write([]byte("User event is of not correct type."))
			return 
		}

		userIdentification, err := uuid.Parse(params.Data.UserID)

		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte(fmt.Sprintf("user ID failed to parse to UUID: %v", err)))
			return
		}

		err = dbQueries.ChirpyRedUpgrade(req.Context(), userIdentification)

		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte(fmt.Sprintf("User account upgrade failed maybe invalid user ID? : %v",err)))
		}

		w.WriteHeader(204)
})




	server.ListenAndServe()
}
