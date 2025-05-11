package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Drnel/chirpy_btdv/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("error acquiring connetion to db")
		return
	}
	dbQueries := database.New(db)
	defer db.Close()

	var serve_mux = http.NewServeMux()
	var server = http.Server{Handler: serve_mux}
	server.Addr = ":8080"
	var apiCfg = apiConfig{dbQueries: dbQueries}

	serve_mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	serve_mux.Handle("GET /admin/metrics", apiCfg.middlewareMetricsShow())
	serve_mux.Handle("POST /admin/reset", apiCfg.middlewareMetricsReset())
	serve_mux.HandleFunc("GET /api/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}))
	serve_mux.HandleFunc("POST /api/users", http.HandlerFunc(apiCfg.addUser()))
	serve_mux.HandleFunc("POST /api/chirps", http.HandlerFunc(apiCfg.addChirp()))
	serve_mux.HandleFunc("GET /api/chirps", http.HandlerFunc(apiCfg.RetrieveChirps()))
	serve_mux.HandleFunc("GET /api/chirps/{chirpID}", http.HandlerFunc(apiCfg.getChirpById()))

	fmt.Println("Starting Chirpy server:")
	server.ListenAndServe()
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsShow() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
		w.Header().Add("Content-type", "text/html")
	})
}

func (cfg *apiConfig) middlewareMetricsReset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("PLATFORM") != "dev" {
			w.WriteHeader(403)
			return
		}
		cfg.dbQueries.ResetUsers(r.Context())
		cfg.fileserverHits.Store(0)

	})
}

func validationHandler(w http.ResponseWriter, r *http.Request) {
	statusCode := 200
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	type returnVals struct {
		Cleaned_body string `json:"cleaned_body"`
	}
	words := strings.Split(params.Body, " ")
	for i, v := range words {
		if strings.ToLower(v) == "kerfuffle" || strings.ToLower(v) == "sharbert" || strings.ToLower(v) == "fornax" {
			words[i] = "****"
		}
	}
	cleaned_sentence := strings.Join(words, " ")

	respBody := returnVals{
		Cleaned_body: cleaned_sentence,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(params.Body) > 140 {
		statusCode = 400

		type returnVals struct {
			Error string `json:"error"`
		}
		respBody := returnVals{
			Error: "Chirp is too long",
		}
		dat, err = json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dat)
}

func (cfg *apiConfig) addUser() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		type parameters struct {
			Email string `json:"email"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}
		user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
		if err != nil {
			log.Printf("Error getting database user registered: %s", err)
			w.WriteHeader(500)
			return
		}
		returnUser := User{}
		returnUser.ID = user.ID
		returnUser.CreatedAt = user.CreatedAt
		returnUser.UpdatedAt = user.UpdatedAt
		returnUser.Email = user.Email

		dat, err := json.Marshal(returnUser)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(dat)
	})
}

func (cfg *apiConfig) addChirp() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		type parameters struct {
			Body    string        `json:"body"`
			User_id uuid.NullUUID `json:"user_id"`
		}
		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}

		if len(params.Body) > 140 {
			type returnVals struct {
				Error string `json:"error"`
			}
			respBody := returnVals{
				Error: "Chirp is too long",
			}
			dat, err := json.Marshal(respBody)
			if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				w.WriteHeader(500)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			w.Write(dat)
			return
		}

		chirp, err := cfg.dbQueries.AddChirp(r.Context(), database.AddChirpParams{
			Body:   params.Body,
			UserID: params.User_id,
		})
		if err != nil {
			log.Printf("Error getting database user registered: %s", err)
			w.WriteHeader(500)
			return
		}
		returnChirp := Chirp{}
		returnChirp.Body = chirp.Body
		returnChirp.CreatedAt = chirp.CreatedAt
		returnChirp.UpdatedAt = chirp.UpdatedAt
		returnChirp.ID = chirp.ID
		returnChirp.User_id = chirp.UserID.UUID

		dat, err := json.Marshal(returnChirp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(dat)

	})
}
func (cfg *apiConfig) RetrieveChirps() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chirps, err := cfg.dbQueries.RetrieveChirps(r.Context())
		if err != nil {
			log.Printf("Error retrieving chirps from database %s", err)
			w.WriteHeader(500)
			return
		}
		returnChirps := []Chirp{}
		for _, chirp := range chirps {
			returnChirp := Chirp{}
			returnChirp.ID = chirp.ID
			returnChirp.CreatedAt = chirp.CreatedAt
			returnChirp.UpdatedAt = chirp.UpdatedAt
			returnChirp.Body = chirp.Body
			returnChirp.User_id = chirp.UserID.UUID
			returnChirps = append(returnChirps, returnChirp)
		}

		dat, err := json.Marshal(returnChirps)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)
	})
}

func (cfg *apiConfig) getChirpById() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			log.Printf("Error converting text to uuid: %s", err)
			w.WriteHeader(500)
			return
		}
		chirps, err := cfg.dbQueries.GetChirpById(r.Context(), id)
		if err != nil {
			log.Printf("Error getting chirp from database by id: **%s**", err)
			w.WriteHeader(500)
			return
		}
		if len(chirps) == 0 {
			w.WriteHeader(404)
			return
		}
		returnChirp := Chirp{}
		returnChirp.ID = chirps[0].ID
		returnChirp.CreatedAt = chirps[0].CreatedAt
		returnChirp.UpdatedAt = chirps[0].UpdatedAt
		returnChirp.Body = chirps[0].Body
		returnChirp.User_id = chirps[0].UserID.UUID

		dat, err := json.Marshal(returnChirp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)
	})
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_id   uuid.UUID `json:"user_id"`
}
