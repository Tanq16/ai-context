package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/ai-context/internal/aicontext"
	"github.com/tanq16/ai-context/internal/server"
)

var serveFlags struct {
	port int
	host string
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch a web server to use the AI Context tool through a UI.",
	Run: func(cmd *cobra.Command, args []string) {
		srv := server.New(serveFlags.host, serveFlags.port)
		if err := srv.Setup(); err != nil {
			log.Fatalf("ERROR [server] Failed to setup: %v", err)
		}
		log.Printf("INFO [server] Starting on %s:%d", serveFlags.host, serveFlags.port)

		srv.RegisterAPI("/generate", generateHandler)
		srv.RegisterAPI("/load", loadHandler)
		srv.RegisterAPI("/clear", clearHandler)
		srv.RegisterAPI("/download", downloadHandler)

		if err := srv.Run(); err != nil {
			log.Fatalf("ERROR [server] Server error: %v", err)
		}
	},
}

type generateRequest struct {
	URL    string   `json:"url"`
	Ignore []string `json:"ignore"`
}

type generateResponse struct {
	Content string `json:"content"`
}


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
	w.WriteHeader(http.StatusNoContent)
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	outputFile, err := findGeneratedFile()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
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
	aicontext.Handler([]string{req.URL}, req.Ignore, 1, true)
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

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}
	mdFile, err := findGeneratedFile()
	if err != nil {
		http.Error(w, "Could not find context file.", http.StatusNotFound)
		log.Printf("Error finding generated file for download: %v", err)
		return
	}
	mdContent, err := os.ReadFile(mdFile)
	if err != nil {
		http.Error(w, "Could not read context file.", http.StatusInternalServerError)
		log.Printf("Error reading context file %s for download: %v", mdFile, err)
		return
	}
	imagesDir := filepath.Join("context", "images")
	images, err := os.ReadDir(imagesDir)
	if err != nil || len(images) == 0 {
		w.Header().Set("Content-Type", "text/markdown")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(mdFile)))
		w.Write(mdContent)
		return
	}
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	mdWriter, err := zipWriter.Create(filepath.Base(mdFile))
	if err != nil {
		http.Error(w, "Failed to create markdown entry in zip.", http.StatusInternalServerError)
		return
	}
	_, err = mdWriter.Write(mdContent)
	if err != nil {
		http.Error(w, "Failed to write markdown content to zip.", http.StatusInternalServerError)
		return
	}
	for _, image := range images {
		if !image.IsDir() {
			imgPath := filepath.Join(imagesDir, image.Name())
			imgData, err := os.ReadFile(imgPath)
			if err != nil {
				log.Printf("Warning: could not read image file %s: %v", imgPath, err)
				continue // Skip this file if it can't be read.
			}
			imgWriter, err := zipWriter.Create(filepath.Join("images", image.Name()))
			if err != nil {
				log.Printf("Warning: could not create image entry %s in zip: %v", image.Name(), err)
				continue
			}
			_, err = imgWriter.Write(imgData)
			if err != nil {
				log.Printf("Warning: could not write image data for %s to zip: %v", image.Name(), err)
			}
		}
	}
	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Failed to finalize zip file.", http.StatusInternalServerError)
		return
	}
	zipName := strings.TrimSuffix(filepath.Base(mdFile), ".md") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))
	w.Write(buf.Bytes())
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
	imgPath := filepath.Join(dir, "images")
	if _, err := os.Stat(imgPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(imgPath); err != nil {
			log.Printf("Failed to remove images directory %s: %v", imgPath, err)
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
	serveCmd.Flags().IntVarP(&serveFlags.port, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVarP(&serveFlags.host, "host", "H", "0.0.0.0", "Host to bind to")
}
