package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var env *Env



// Env will hold our application's dependencies
type Env struct {
	db *sql.DB
}

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// hanlde what happen when someone visit our server's root url*
// adding (env *Env) make handlePostsRequest a method on the *Env type
func (env *Env) handlePostsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := r.URL.Path

		if path == "/posts/" {
			rows, err := env.db.Query("SELECT id, title, content, created_at, updated_at FROM posts")

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Close the rows when we are done to free up the db connection
			defer rows.Close()

			var dbPosts []Post

			for rows.Next() {
				var post Post
				err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				dbPosts = append(dbPosts, post)

			}
			dbPostsJSON, err := json.Marshal(dbPosts)

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(dbPostsJSON)

		} else {
			id, err := strconv.Atoi(strings.TrimPrefix(path, "/posts/"))

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			row := env.db.QueryRow("SELECT id, title, content, created_at, updated_at FROM posts WHERE id = $1", id)

			var post Post

			err = row.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt)

			if err != nil {
				if err == sql.ErrNoRows {
					http.NotFound(w, r)
					return
				}

				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			postJSON, err := json.Marshal(post)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(postJSON)

		}

	case http.MethodPost:
		var newID int
		var newPost Post

		err := json.NewDecoder(r.Body).Decode(&newPost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sqlStatement := `INSERT INTO  posts (title, content, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`

		err = env.db.QueryRow(sqlStatement, newPost.Title, newPost.Content, newPost.CreatedAt, newPost.UpdatedAt).Scan(&newID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newPost.ID = newID
		newPostJSON, err := json.Marshal(newPost)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(newPostJSON)

	case http.MethodPut:
		path := r.URL.Path

		id, err := strconv.Atoi(strings.TrimPrefix(path, "/posts/"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var updatedPost Post

		err = json.NewDecoder(r.Body).Decode(&updatedPost)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := env.db.Exec("UPDATE posts SET title = $1, content = $2, updated_at = $3 WHERE id = $4", updatedPost.Title, updatedPost.Content, time.Now(), id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rowAffected, err := result.RowsAffected(); rowAffected == 0 {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowOfUpdatedPost := env.db.QueryRow("SELECT * FROM posts WHERE id = $1", id)

		err = rowOfUpdatedPost.Scan(&updatedPost.ID, &updatedPost.Title, &updatedPost.Content, &updatedPost.CreatedAt, &updatedPost.UpdatedAt)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		updatedPostJSON, err := json.Marshal(updatedPost)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(updatedPostJSON)

	case http.MethodDelete:
		path := r.URL.Path

		id, err := strconv.Atoi(strings.TrimPrefix(path, "/posts/"))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result, err := env.db.Exec("DELETE FROM posts WHERE id = $1", id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if rowAffected, err := result.RowsAffected(); rowAffected == 0 {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	}
}

// CORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the necessary CORS headers
		// Use '*' for development. In production, restrict this to front end's domain
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// If it's an OPTIONS requests, just send the headers and a 200 OK
		if r.Method == "OPTIONS" {
			return
		}

		// Otherwise, call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// init() runs once when the serverless function is started ("cold start")
func init() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL environment variable is not set")
    }

    db, err := sql.Open("pgx", dbURL)
    if err != nil {
        log.Fatalf("failed to open db %s: %s", dbURL, err)
    }
    
    // We can't use `defer db.Close()` in init(), as the function exits immediately.
    // The serverless environment will manage the connection's lifecycle.

    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    createTableSQL := `CREATE TABLE IF NOT EXISTS posts (
        id SERIAL PRIMARY KEY,
        title TEXT,
        content TEXT,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW()
    );`
    
    if _, err := db.Exec(createTableSQL); err != nil {
        log.Fatal(err)
    }

    // Initialize our global env variable
    env = &Env{db: db}
    fmt.Println("Database connection and table check successful.")
}

// Handler is the exported function that Vercel will run for each request.
// It must have this exact signature.
func Handler(w http.ResponseWriter, r *http.Request) {
    // We are simply using our existing handler method.
    // We create a standard http.Handler from our method, wrap it in our
    // CORS middleware, and then call its ServeHTTP method.
    
    router := http.HandlerFunc(env.handlePostsRequest)
    corsWrapper := enableCORS(router)
    
    corsWrapper.ServeHTTP(w, r)
}
