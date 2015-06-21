package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var (
	envUsername string
	envPassword string
	envRoot     string
	envPort     string
	envHost     string
	envPath     string
)

func init() {
	envUsername = os.Getenv("GYOZO_USERNAME")
	envPassword = os.Getenv("GYOZO_PASSWORD")
	envRoot = os.Getenv("GYOZO_ROOT")
	envPort = os.Getenv("GYOZO_PORT")
	envHost = os.Getenv("GYOZO_HOST")
	envPath = os.Getenv("GYOZO_PATH")

	rand.Seed(time.Now().UnixNano())
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/upload", uploadHandler).Methods("POST")
	r.HandleFunc("/{id:[0-9a-f]{8}}", imageHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(envHost+":"+envPort, nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, http.StatusText(401), 401)
		return
	}
	if !(username == envUsername && password == envPassword) {
		http.Error(w, http.StatusText(403), 403)
		return
	}
	f, _, err := r.FormFile("imagedata")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		http.Error(w, http.StatusText(415), 415)
		return
	}
	var fn string
	var file *os.File
	for file == nil {
		fn = fmt.Sprintf("%08x", rand.Uint32())
		file, _ = os.Create(envPath + "/" + fn + ".png")
	}
	defer file.Close()
	png.Encode(file, img)
	http.Redirect(w, r, envRoot+"/"+fn, 302)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	f, err := os.Open(envPath + "/" + id + ".png")
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	defer f.Close()
	w.Header().Add("Content-Type", "image/png")
	io.Copy(w, f)
}
