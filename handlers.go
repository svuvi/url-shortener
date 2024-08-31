package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
)

type writeRequest struct {
	LongUrl string `json:"longUrl"`
}

type writeResponse struct {
	ShortUrl string `json:"shortUrl"`
}

type accessCountResponse struct {
	LongUrl     string `json:"longUrl"`
	AccessCount int    `json:"accessCount"`
}

func (c Context) writeService(w http.ResponseWriter, r *http.Request) {
	writeRequest := writeRequest{}
	err := decodeJSONBody(w, r, &writeRequest)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	_, err = url.ParseRequestURI(writeRequest.LongUrl)
	if err != nil {
		http.Error(w, "URL string is invalid", http.StatusBadRequest)
	}

	result, err := c.db.Exec("INSERT INTO urls (original_url, access_count) VALUES (?, ?)", writeRequest.LongUrl, 0)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeResponse := writeResponse{
		ShortUrl: base62Encode(int(insertedID)),
	}

	jsonResponse, err := json.Marshal(writeResponse)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (c Context) redirectService(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.PathValue("shortUrl")
	if len(shortUrl) > 6 {
		http.Error(w, "This short link is longer than it is allowed to be. There can't be more than 6 symbols after \"/\"", http.StatusBadRequest)
		return
	}

	id, err := base62Decode(shortUrl)
	if err != nil {
		http.Error(w, "There are invalid characters in the shortened link", http.StatusBadRequest)
		return
	}

	var longUrl string
	err = c.db.QueryRow("SELECT original_url FROM urls WHERE id=?", id).Scan(&longUrl)
	if err == sql.ErrNoRows {
		http.Error(w, "This link doesn't exist", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longUrl, http.StatusFound)

	_, err = c.db.Exec("UPDATE urls SET access_count = access_count + 1 WHERE id=?", id)
	if err != nil {
		log.Print("Failed to increment the access_count:\n", err)
	}
}

func (c Context) accessCount(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.PathValue("shortUrl")
	if len(shortUrl) > 6 {
		http.Error(w, "This short link is longer than it is allowed to be. There can't be more than 6 symbols after \"/\"", http.StatusBadRequest)
		return
	}

	id, err := base62Decode(shortUrl)
	if err != nil {
		http.Error(w, "There are invalid characters in the shortened link", http.StatusBadRequest)
		return
	}

	var a accessCountResponse
	err = c.db.QueryRow("SELECT original_url, access_count FROM urls WHERE id=?", id).Scan(&a.LongUrl, &a.AccessCount)
	if err == sql.ErrNoRows {
		http.Error(w, "This link doesn't exist", http.StatusNotFound)
		return
	}

	jsonResponse, err := json.Marshal(a)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
