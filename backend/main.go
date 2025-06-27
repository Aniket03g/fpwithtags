package main

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const outDir = "../frontend/out"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Request] %s %s", r.Method, r.URL.Path)
		// Gzip response if supported
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
			serveStatic(gzw, r)
		} else {
			serveStatic(w, r)
		}
	})

	log.Println("Serving on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("[serveStatic] Path: %s", path)
	// Serve static JSON files from /data
	if strings.HasPrefix(path, "/data/") {
		filePath := filepath.Join(outDir, path)
		log.Printf("[serveStatic] Serving data file: %s", filePath)
		if fileExists(filePath) {
			http.ServeFile(w, r, filePath)
			return
		}
		log.Printf("[serveStatic] Data file not found: %s", filePath)
		serve404(w, r)
		return
	}

	// Serve static assets if they exist
	filePath := filepath.Join(outDir, path)
	if fileExists(filePath) && !isDir(filePath) {
		log.Printf("[serveStatic] Serving static file: %s", filePath)
		http.ServeFile(w, r, filePath)
		return
	}

	// Serve 404.html for /404 or missing files
	if path == "/404" || path == "/404.html" {
		log.Printf("[serveStatic] Serving 404 page")
		serve404(w, r)
		return
	}

	// For all other routes, serve index.html (SPA fallback)
	log.Printf("[serveStatic] Fallback to index.html for path: %s", path)
	http.ServeFile(w, r, filepath.Join(outDir, "index.html"))
}

func serve404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	log.Printf("[serve404] Serving 404.html for path: %s", r.URL.Path)
	http.ServeFile(w, r, filepath.Join(outDir, "404.html"))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
