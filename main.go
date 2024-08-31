package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Context struct {
	db *sql.DB
}

const dbSchema = `
CREATE TABLE IF NOT EXISTS urls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	original_url TEXT NOT NULL,
	access_count INTEGER NOT NULL
);
`

func main() {
	db, err := sql.Open("sqlite3", "./urls.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(dbSchema)
	if err != nil {
		log.Fatal(err)
	}

	c := Context{
		db: db,
	}

	http.HandleFunc("GET /{shortUrl}", c.redirectService)
	http.HandleFunc("POST /api/url/shorten", c.writeService)
	http.HandleFunc("GET /api/url/access_count/{shortUrl}", c.accessCount)

	log.Printf("Serving: http://localhost:3000/")
	err = http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("Unable to start the server: ", err)
	}
}
