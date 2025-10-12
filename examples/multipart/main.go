package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
)

// FileUploadRequest for file uploads
type FileUploadRequest struct {
	Title       string                 `multipart:"title"`
	Description string                 `multipart:"description"`
	File        *decoder.MultipartFile `multipart:"file"`
	AllFiles    decoder.MultipartFiles `multipart:",allfiles"`
}

// UploadResponse represents the API response
type UploadResponse struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	FileName    string   `json:"file_name"`
	FileSize    int64    `json:"file_size"`
	TotalFiles  int      `json:"total_files"`
	AllFiles    []string `json:"all_files"`
}

func main() {
	// Initialize roamer with multipart decoder
	r := roamer.NewRoamer(
		roamer.WithDecoders(
			decoder.NewMultipartFormData(
				decoder.WithMaxMemory(64<<20), // 64MB
			),
		),
		roamer.WithParsers(parser.NewQuery()),
	)

	http.HandleFunc("/upload", func(w http.ResponseWriter, req *http.Request) {
		var uploadReq FileUploadRequest

		if err := r.Parse(req, &uploadReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Process uploaded file
		var fileName string
		var fileSize int64
		if uploadReq.File != nil && uploadReq.File.Header != nil {
			fileName = uploadReq.File.Header.Filename
			fileSize = uploadReq.File.Header.Size
			// Here you would typically save the file
			// file, _ := uploadReq.File.Header.Open()
			// defer file.Close()
		}

		// Get all file names
		var allFileNames []string
		for _, f := range uploadReq.AllFiles {
			if f.Header != nil {
				allFileNames = append(allFileNames, f.Header.Filename)
			}
		}

		response := UploadResponse{
			Title:       uploadReq.Title,
			Description: uploadReq.Description,
			FileName:    fileName,
			FileSize:    fileSize,
			TotalFiles:  len(uploadReq.AllFiles),
			AllFiles:    allFileNames,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/upload \\")
	log.Println("  -F 'title=My Document' \\")
	log.Println("  -F 'description=Important file' \\")
	log.Println("  -F 'file=@/path/to/your/file.txt'")
	log.Println()
	log.Println("Or create a test file and upload:")
	log.Println("  echo 'test content' > test.txt")
	log.Println("  curl -X POST http://localhost:8080/upload \\")
	log.Println("    -F 'title=Test' \\")
	log.Println("    -F 'description=Test file' \\")
	log.Println("    -F 'file=@test.txt'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
