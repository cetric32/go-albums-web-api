package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Album struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []Album{
	{ID: 1, Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: 2, Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: 3, Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func getAlbumsHandler(resp http.ResponseWriter, req *http.Request) {
	// Convert the albums variable to JSON
	albumsJson, err := json.Marshal(albums)
	if err != nil {
		// If there is an error, print a message and return a status code of 500
		writeResponseError(resp, http.StatusInternalServerError, err.Error())
		return
	}

	// If all goes well, write the JSON list of albums to the response
	writeResponse(resp, albumsJson)
}

func getAlbumHandler(resp http.ResponseWriter, req *http.Request) {
	// Get the id from the request
	id, error := strconv.Atoi(req.URL.Query().Get("id"))
	if error != nil {
		// If there is an error, print a message and return a status code of 400
		writeResponseError(resp, http.StatusBadRequest, "id is not valid")
		return
	}

	// Loop through the list of albums, looking for an album with the given id
	for _, album := range albums {
		if id == album.ID {
			// Convert the album variable to JSON
			albumJson, err := json.Marshal(album)
			if err != nil {
				// If there is an error, print a message and return a status code of 500
				writeResponseError(resp, http.StatusInternalServerError, err.Error())
				return
			}

			// If all goes well, write the JSON list of albums to the response
			writeResponse(resp, albumJson)
			return
		}
	}

	// If no album is found, return a 404 status code
	writeResponseError(resp, http.StatusNotFound, "album not found")
}

func postAlbumHandler(resp http.ResponseWriter, req *http.Request) {
	// check if post request
	if req.Method != http.MethodPost {
		writeResponseError(resp, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Decode the request body into a new Album instance
	var newAlbum Album
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&newAlbum); err != nil {
		// If there is an error, print a message and return a status code of 400
		writeResponseError(resp, http.StatusBadRequest, "invalid request payload")
		return
	}

	// check if id already exists and throw error
	for _, album := range albums {
		if newAlbum.ID == album.ID {
			writeResponseError(resp, http.StatusConflict, "id already exists")
			return
		}
	}

	// Append the new album to the list of albums
	albums = append(albums, newAlbum)

	// Convert the new album variable to JSON
	albumsJson, err := json.Marshal(albums)
	if err != nil {
		// If there is an error, print a message and return a status code of 500
		writeResponseError(resp, http.StatusInternalServerError, err.Error())
		return
	}

	// If all goes well, write the JSON list of albums to the response
	writeResponse(resp, albumsJson)
}

func writeResponseError(resp http.ResponseWriter, status int, message string) {
	jsonError, _ := json.Marshal(map[string]string{"error": message})
	resp.Header().Set("content-type", "application/json")
	resp.WriteHeader(status)
	resp.Write(jsonError)
}

func writeResponse(resp http.ResponseWriter, data []byte) {
	// Set the content type header on the response to application/json
	resp.Header().Set("content-type", "application/json")
	// resp.WriteHeader(status)
	resp.Write(data)
}

func main() {
	http.HandleFunc("/albums", getAlbumsHandler)
	http.HandleFunc("/album", getAlbumHandler)
	http.HandleFunc("/newAlbum", postAlbumHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
