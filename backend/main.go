package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

			row := env.db.QueryRow("SELECT id, title, content, created_at, updated_at FROM posts WHERE id = ?", id)

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
		var newPost Post

		// Read data from incoming request body
		err := json.NewDecoder(r.Body).Decode(&newPost)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newPost.CreatedAt = time.Now()
		newPost.UpdatedAt = time.Now()

		sqlStatement := `INSERT INTO posts (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`

		result, err := env.db.Exec(sqlStatement, newPost.Title, newPost.Content, newPost.CreatedAt, newPost.UpdatedAt)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newID, err := result.LastInsertId()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newPost.ID = int(newID)

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

		result, err := env.db.Exec("UPDATE posts SET title = ?, content = ?, updated_at = ? WHERE id = ?", updatedPost.Title, updatedPost.Content, time.Now(), id)

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

		rowOfUpdatedPost := env.db.QueryRow("SELECT * FROM posts WHERE id = ?", id)

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

		result, err := env.db.Exec("DELETE FROM posts WHERE id = ?", id)

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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
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

func main() {
	db, err := sql.Open("sqlite3", "./blog.db")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Dabase connection successful.")

	createTableSQL := `CREATE TABLE IF NOT EXISTS posts (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"title" TEXT,
		"content" TEXT,
		"created_at" DATETIME,
		"updated_at" DATETIME
	);`

	env := &Env{db: db}

	_, err = db.Exec(createTableSQL)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Posts table created or already exists")

	fs := http.FileServer(http.Dir("./frontend"))

	http.Handle("/posts/", enableCORS(http.HandlerFunc(env.handlePostsRequest)))

	http.Handle("/", fs)

	fmt.Println("Server is listening on port 8080...")

	// http.ListenAndServe Start the server at port 8080
	// Wrap with log.Fatal so if the server failed to start, the program exit with error
	log.Fatal(http.ListenAndServe(":8080", nil))
}
