package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fasthttp/router"
	_ "github.com/mattn/go-sqlite3"
	"github.com/valyala/fasthttp"
)

var db *sql.DB

// copy pasta
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./jokes.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create the jokes table if it doesn't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS jokes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        content TEXT NOT NULL
    );`
	if _, err := db.Exec(createTableQuery); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func pingHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain; charset=utf8")
	currentTime := time.Now().Format(time.RFC1123)
	fmt.Fprintf(ctx, "Hello there General! %s", currentTime)
}

// GET /jokes - Returns the list of jokes in JSON format
func getJokesHandler(ctx *fasthttp.RequestCtx) {
	rows, err := db.Query("SELECT content FROM jokes")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "Database error: %v", err)
		return
	}
	defer rows.Close()

	var jokes []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			fmt.Fprintf(ctx, "Error reading joke: %v", err)
			return
		}
		jokes = append(jokes, content)
	}

	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(jokes); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "Error encoding response: %v", err)
	}
}

// POST /jokes - Accepts a JSON object and adds a new joke to the database
func postJokeHandler(ctx *fasthttp.RequestCtx) {
	var joke map[string]string
	if err := json.Unmarshal(ctx.PostBody(), &joke); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		fmt.Fprintf(ctx, "Invalid JSON")
		return
	}

	jokeContent, exists := joke["joke"]
	if !exists || jokeContent == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		fmt.Fprintf(ctx, "Joke content cannot be empty")
		return
	}

	_, err := db.Exec("INSERT INTO jokes (content) VALUES (?)", jokeContent)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "Database error: %v", err)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"message": "Created"}`)
}

func main() {
	// Initialize the database
	initDB()
	defer db.Close()

	// Set up router and define routes
	r := router.New()
	r.GET("/ping", pingHandler)
	r.GET("/jokes", getJokesHandler)
	r.POST("/jokes", postJokeHandler)

	// Start the server on port 8080
	log.Println("Server running on port 8080")
	if err := fasthttp.ListenAndServe(":8080", r.Handler); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
