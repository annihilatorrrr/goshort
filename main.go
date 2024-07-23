package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jxskiss/base62"
)

type requestBody struct {
	Hash string `json:"hash"`
	Err  string `json:"err"`
}

type rrequestBody struct {
	URL string `json:"url"`
}

var key = []byte("annianniannianni")

func handleico(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed!", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./favicon.ico")
}

func EncodeBase62(data []byte) string {
	return string(base62.Encode(data))
}

func DecodeBase62(encoded string) ([]byte, error) {
	return base62.Decode([]byte(encoded))
}

func dowork(r *http.Request) *requestBody {
	jdata := &requestBody{
		Hash: "",
		Err:  "",
	}
	var rrrr rrequestBody
	if err := json.NewDecoder(r.Body).Decode(&rrrr); err != nil {
		jdata.Err = "Error decoding JSON: " + err.Error()
		return jdata
	}
	longURL := rrrr.URL
	if longURL == "" {
		jdata.Err = "URL parameter is missing!"
		return jdata
	}
	shortURL, err := encodeURL([]byte(longURL))
	if err != nil {
		jdata.Err = "Error encoding URL: " + err.Error()
	} else {
		jdata.Hash = shortURL
	}
	return jdata
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed!", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dowork(r)); err != nil {
		http.Error(w, "Error encoding JSON response: "+err.Error(), http.StatusInternalServerError)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed!", http.StatusMethodNotAllowed)
		return
	}
	shortCode := strings.TrimPrefix(r.URL.Path, "/r/")
	if shortCode == "" {
		http.Error(w, "Short code is missing!", http.StatusBadRequest)
		return
	}
	longURL, err := decodeURL(shortCode)
	if err != nil {
		http.Error(w, "Invalid short code: "+err.Error(), http.StatusBadRequest)
		return
	}
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta http-equiv="refresh" content="3;url=` + longURL + `" />
	</head>
	<body>
		<p>Please wait 3 seconds ...</p>
	</body>
	</html>`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(htmlContent))
}

func encodeURL(plaintext []byte) (ciphertext string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return EncodeBase62(gcm.Seal(nonce, nonce, plaintext, nil)), nil
}

func decodeURL(ciphertext string) (plaintext string, err error) {
	ciphertextBytes, err := DecodeBase62(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertextBytes := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	plainTextBytes, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plainTextBytes), nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed!", http.StatusMethodNotAllowed)
		return
	}
	indexPath := filepath.Join(".", "index.html")
	http.ServeFile(w, r, indexPath)
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/sort", shortenHandler)
	router.HandleFunc("/r/", redirectHandler)
	router.HandleFunc("/favicon.ico", handleico)
	httphandler := http.Handler(router)
	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))
	port := os.Getenv("PORT")
	if port == "" {
		port = "9097"
	}
	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Handler:      httphandler,
	}
	go func() {
		time.Sleep(time.Second * 21600)
		self, err := os.Executable()
		if err != nil {
			log.Println(err.Error())
			return
		}
		_ = syscall.Exec(self, os.Args, os.Environ())
	}()
	log.Println("Started!")
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
	log.Fatal("Bye!")
}
