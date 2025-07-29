package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tanq16/ai-context/aicontext"
)

//go:embed all:web
var webFS embed.FS

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch a web server to use the AI Context tool through a UI.",
	Run:   runServer,
}

type generateRequest struct {
	URL    string   `json:"url"`
	Ignore []string `json:"ignore"`
}

type generateResponse struct {
	Content string `json:"content"`
}

func runServer(cmd *cobra.Command, args []string) {
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Failed to create static file system: %v", err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/load", loadHandler)
	http.HandleFunc("/clear", clearHandler) // Add the new /clear route
	http.HandleFunc("/", rootHandler(staticFS))

	port := "8080"
	fmt.Printf("Starting server at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func rootHandler(fs fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.FileServer(http.FS(fs)).ServeHTTP(w, r)
			return
		}
		indexHTML, err := webFS.ReadFile("web/index.html")
		if err != nil {
			http.Error(w, "Could not read index.html", http.StatusInternalServerError)
			log.Printf("Error reading embedded index.html: %v", err)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	}
}

// clearHandler handles the request to delete the generated context file.
func clearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := cleanupContextDir(); err != nil {
		log.Printf("Error during cleanup: %v", err)
		http.Error(w, "Failed to clear context file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Success, no content to return
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	outputFile, err := findGeneratedFile()
	if err != nil {
		json.NewEncoder(w).Encode(generateResponse{Content: ""})
		return
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		http.Error(w, "Failed to read context file", http.StatusInternalServerError)
		log.Printf("Error reading output file %s: %v", outputFile, err)
		return
	}

	resp := generateResponse{Content: string(content)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if err := cleanupContextDir(); err != nil {
		log.Printf("Warning: could not clean up context directory: %v", err)
	}

	aicontext.Handler([]string{req.URL}, req.Ignore, 1)

	outputFile, err := findGeneratedFile()
	if err != nil {
		http.Error(w, "Failed to find generated context file", http.StatusInternalServerError)
		log.Printf("Error finding generated file: %v", err)
		return
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		http.Error(w, "Failed to read context file", http.StatusInternalServerError)
		log.Printf("Error reading output file %s: %v", outputFile, err)
		return
	}

	resp := generateResponse{Content: string(content)}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func cleanupContextDir() error {
	dir := "context"
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			log.Printf("Failed to remove file %s: %v", file, err)
		}
	}
	return nil
}

func findGeneratedFile() (string, error) {
	dir := "context"
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return "", fmt.Errorf("error searching for files: %w", err)
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no markdown file found in context directory")
	}
	return files[0], nil
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
