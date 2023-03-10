package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const uploads = "/etc/file-server/uploads"
const success = `<!DOCTYPE html>
<html>
  <head>
    <title>Upload Complete</title>
  </head>
  <body>
    <p>Upload successful.</p>
  </body>
</html>
`

// decompose takes a base directory and a destination path relative to the
// base, and returns the fully qualified destination including the base, except
// that relative path parts (i.e. "..") have been removed. This means that the
// resulting paths will always be descendants of base (no jail-breaking, if we
// ignore links).
func decompose(base, destination string) string {
	parts := strings.Split(destination, string(os.PathSeparator))
	safeParts := make([]string, len(parts))
	for _, part := range parts {
		if part != "." && part != ".." {
			safeParts = append(safeParts, part)
		}
	}
	safePath := strings.Join(safeParts, string(os.PathSeparator))
	return filepath.Join(base, safePath)
}

func main() {
	http.HandleFunc("/", handleRequest)

	listen := ":80"
	log.Printf("About to listen on %v", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	destination := r.Header.Get("X-Destination-File")
	if destination == "" {
		deliverError(w, 400, "X-Destination-File header missing from request")
		return
	}
	destination, err := url.PathUnescape(destination)
	if err != nil {
		deliverError(w, 400, "Invalid destination file name")
		return
	}
	multi, err := r.MultipartReader()
	if err != nil {
		deliverError(w, 400, "Request body is not multipart")
		return
	}
	part, err := multi.NextPart()
	if err != nil {
		deliverError(w, 400, "Request body does not contain a multipart piece")
		return
	}
	destination = decompose(uploads, destination)
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		deliverError(w, 500, "Unable to create parent directory of destination file")
		return
	}
	file, err := os.Create(destination)
	if err != nil {
		deliverError(w, 500, "Unable to create destination file")
		return
	}
	defer func() {
		file.Sync()
		file.Close()
	}()
	_, err = io.Copy(file, part)
	if err != nil {
		deliverError(w, 500, "Unable to copy request file into destination file")
		return
	}
	deliverSuccess(w)
}

func deliverError(w http.ResponseWriter, status int, message string) {
	log.Print(message)
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(status)
	if _, err := io.WriteString(w, message); err != nil {
		log.Print(err)
	}
}

func deliverSuccess(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	if _, err := io.WriteString(w, success); err != nil {
		log.Print(err)
	}
}
