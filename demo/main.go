package main

import (
	"bytes"
	"github.com/motoki317/go-identicon"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	args := strings.Split(r.URL.Path, "/")
	args = args[1:]

	if len(args) != 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	item := args[0]

	// support jpg too?
	if !strings.HasSuffix(item, ".png") {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	item = strings.TrimSuffix(item, ".png")

	code := identicon.Code(item)
	size := 1024
	settings := identicon.DefaultSettings()

	//log.Printf("got settings '%s'\n", settings)

	img, err := identicon.Render(code, size, settings)
	if err != nil {
		log.Println("unable to render image:", err)
		return
	}

	log.Printf("creating identicon for '%s'\n", item)

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}

func main() {
	port := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}
	log.Printf("Listening on http://localhost%s\n", port)

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(port, nil))
}
