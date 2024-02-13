package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Album struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// var albums = []Album{
// 	{ID: 1, Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
// 	{ID: 2, Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
// 	{ID: 3, Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
// }

var dbHost string
var dbUser string
var dbPass string
var dbName string
var dbPort string

var dbHandle *sql.DB

func getAlbumsHandler(resp http.ResponseWriter, req *http.Request) {

	sqlStatement := "SELECT * FROM albums"

	rows, error := dbHandle.Query(sqlStatement)

	if error != nil {

		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

	defer rows.Close()

	var albums []Album

	for rows.Next() {
		var album Album
		error = rows.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

		if error != nil {
			writeResponseError(resp, http.StatusInternalServerError, error.Error())
			return
		}

		albums = append(albums, album)
	}

	error = rows.Err()

	if error != nil {
		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

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

	// Check if the id is valid
	if id < 1 {
		// If there is an error, print a message and return a status code of 400
		writeResponseError(resp, http.StatusBadRequest, "id is not valid")
		return
	}

	sqlStatement := "SELECT * FROM albums WHERE id = ?"

	row := dbHandle.QueryRow(sqlStatement, id)

	var album Album

	error = row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

	if error != nil {
		errorMessage := error.Error()
		if errorMessage == "sql: no rows in result set" {
			writeResponseError(resp, http.StatusNotFound, "album not found")
		} else {
			writeResponseError(resp, http.StatusInternalServerError, error.Error())
		}

		return
	}

	// Convert the album variable to JSON
	albumJson, err := json.Marshal(album)

	if err != nil {
		// If there is an error, print a message and return a status code of 500
		writeResponseError(resp, http.StatusInternalServerError, err.Error())
		return
	}

	// If all goes well, write the JSON list of albums to the response
	writeResponse(resp, albumJson)

}

func postAlbumHandler(resp http.ResponseWriter, req *http.Request) {
	// check if post request
	if req.Method != http.MethodPost {
		writeResponseError(resp, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// check if content type is application/json
	if req.Header.Get("Content-Type") != "application/json" {
		writeResponseError(resp, http.StatusUnsupportedMediaType, "unsupported media type")
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

	// check if title already exist

	sqlStatement := "SELECT * FROM albums WHERE title = ?"

	row := dbHandle.QueryRow(sqlStatement, newAlbum.Title)

	var album Album

	error := row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

	if error == nil {
		writeResponseError(resp, http.StatusConflict, "title already exists")
		return
	}

	// Insert the new album into the database
	sqlStatement = "INSERT INTO albums (title, artist, price) VALUES (?, ?, ?)"

	_, error = dbHandle.Exec(sqlStatement, newAlbum.Title, newAlbum.Artist, newAlbum.Price)

	if error != nil {
		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

	// Get the list of albums from the database
	sqlStatement = "SELECT * FROM albums"

	rows, error := dbHandle.Query(sqlStatement)

	if error != nil {
		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

	defer rows.Close()

	var albums []Album

	for rows.Next() {
		var album Album
		error = rows.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

		if error != nil {
			writeResponseError(resp, http.StatusInternalServerError, error.Error())
			return
		}

		albums = append(albums, album)

	}

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

func editAlbumHandler(resp http.ResponseWriter, req *http.Request) {
	// check if put request
	if req.Method != http.MethodPut {
		writeResponseError(resp, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// check if content type is application/json
	if req.Header.Get("Content-Type") != "application/json" {
		writeResponseError(resp, http.StatusUnsupportedMediaType, "unsupported media type")
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

	// check if id is valid
	if newAlbum.ID < 1 {
		writeResponseError(resp, http.StatusBadRequest, "id is not valid")
		return
	}

	// check if album exist
	sqlStatement := "SELECT * FROM albums WHERE id = ?"

	row := dbHandle.QueryRow(sqlStatement, newAlbum.ID)

	var album Album

	error := row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

	if error != nil {
		errorMessage := error.Error()
		if errorMessage == "sql: no rows in result set" {
			writeResponseError(resp, http.StatusNotFound, "album not found")
		} else {
			writeResponseError(resp, http.StatusInternalServerError, error.Error())
		}

		return
	}

	if album.Title != newAlbum.Title {
		// check if title already exist
		sqlStatement = "SELECT * FROM albums WHERE title = ?"

		row = dbHandle.QueryRow(sqlStatement, newAlbum.Title)

		error = row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

		if error == nil {
			writeResponseError(resp, http.StatusConflict, "title already exists")
			return
		}
	}

	// Update the album in the database
	if newAlbum.Title == "" {
		newAlbum.Title = album.Title
	}

	if newAlbum.Artist == "" {
		newAlbum.Artist = album.Artist
	}

	if newAlbum.Price == 0 {
		newAlbum.Price = album.Price
	}

	// Update the album in the database
	sqlStatement = "UPDATE albums SET title = ?, artist = ?, price = ? WHERE id = ?"

	_, error = dbHandle.Exec(sqlStatement, newAlbum.Title, newAlbum.Artist, newAlbum.Price, newAlbum.ID)

	if error != nil {
		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

	// Get the list of albums from the database
	sqlStatement = "SELECT * FROM albums"

	rows, error := dbHandle.Query(sqlStatement)

	if error != nil {
		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

	defer rows.Close()

	var albums []Album

	for rows.Next() {
		var album Album
		error = rows.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

		if error != nil {
			writeResponseError(resp, http.StatusInternalServerError, error.Error())
			return
		}

		albums = append(albums, album)

	}

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

func deleteAlbumHandler(resp http.ResponseWriter, req *http.Request) {
	// check if delete request
	if req.Method != http.MethodDelete {
		writeResponseError(resp, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Get the id from the request
	id, error := strconv.Atoi(req.URL.Query().Get("id"))
	if error != nil {
		// If there is an error, print a message and return a status code of 400
		writeResponseError(resp, http.StatusBadRequest, "id is not valid")
		return
	}

	// Check if the id is valid
	if id < 1 {
		// If there is an error, print a message and return a status code of 400
		writeResponseError(resp, http.StatusBadRequest, "id is not valid")
		return
	}

	// check if album exist
	sqlStatement := "SELECT * FROM albums WHERE id = ?"

	row := dbHandle.QueryRow(sqlStatement, id)

	var album Album

	error = row.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

	if error != nil {
		errorMessage := error.Error()
		if errorMessage == "sql: no rows in result set" {
			writeResponseError(resp, http.StatusNotFound, "album not found")
		} else {
			writeResponseError(resp, http.StatusInternalServerError, error.Error())
		}

		return
	}

	// Delete the album from the database
	sqlStatement = "DELETE FROM albums WHERE id = ?"

	_, error = dbHandle.Exec(sqlStatement, id)

	if error != nil {
		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

	// Get the list of albums from the database
	sqlStatement = "SELECT * FROM albums"

	rows, error := dbHandle.Query(sqlStatement)

	if error != nil {
		writeResponseError(resp, http.StatusInternalServerError, error.Error())
		return
	}

	defer rows.Close()

	var albums []Album

	for rows.Next() {
		var album Album
		error = rows.Scan(&album.ID, &album.Title, &album.Artist, &album.Price)

		if error != nil {
			writeResponseError(resp, http.StatusInternalServerError, error.Error())
			return
		}

		albums = append(albums, album)

	}

	// Convert the albums variable to JSON
	albumsJson, err := json.Marshal(albums)

	if err != nil {
		// If there is an
		// error, print a message and return a status code of 500
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
	error := godotenv.Load()

	if error != nil {
		log.Fatal("Error loading .env file")
	}

	// Get connection parameters from environment variables
	dbHost = os.Getenv("DB_HOST")
	dbUser = os.Getenv("DB_USER")
	dbPass = os.Getenv("DB_PASS")
	dbName = os.Getenv("DB_NAME")
	dbPort = os.Getenv("DB_PORT")

	dsn := dbUser + ":" + dbPass + "@tcp(" +
		dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"

	dbHandle, error = sql.Open("mysql", dsn)

	if error != nil {
		log.Fatal("Error connecting to the database:", error)
	}

	defer dbHandle.Close()

	// Check if the connection to the database is working
	error = dbHandle.Ping()
	if error != nil {
		log.Fatal("Error pinging the database:", error)
	}

	// Print a message to the console to confirm the connection
	println("Connected to the database")

	http.HandleFunc("/albums", getAlbumsHandler)
	http.HandleFunc("/album", getAlbumHandler)
	http.HandleFunc("/newAlbum", postAlbumHandler)
	http.HandleFunc("/editAlbum", editAlbumHandler)
	http.HandleFunc("/deleteAlbum", deleteAlbumHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
