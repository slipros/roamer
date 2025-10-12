package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/slipros/roamer"
)

// YAMLDecoder implements a custom decoder for YAML content type.
// It demonstrates how to add support for content types not included
// in roamer by default.
type YAMLDecoder struct{}

// NewYAMLDecoder creates a new YAML decoder.
func NewYAMLDecoder() *YAMLDecoder {
	return &YAMLDecoder{}
}

// ContentType returns the content type that this decoder handles.
// This method is required by the roamer.Decoder interface.
func (d *YAMLDecoder) ContentType() string {
	return "application/yaml"
}

// Tag returns the struct tag name for YAML fields.
// This method is required by the roamer.Decoder interface.
func (d *YAMLDecoder) Tag() string {
	return "yaml"
}

// Decode reads the request body and unmarshals it as YAML into the destination.
// This method is required by the roamer.Decoder interface.
func (d *YAMLDecoder) Decode(r *http.Request, dest any) error {
	// Read the entire body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	// Unmarshal YAML into the destination
	if err := yaml.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return nil
}

// Config represents a typical configuration structure.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Features FeatureFlags   `yaml:"features"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  bool   `yaml:"tls"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type FeatureFlags struct {
	EnableCache   bool `yaml:"enable_cache"`
	EnableMetrics bool `yaml:"enable_metrics"`
	DebugMode     bool `yaml:"debug_mode"`
}

func main() {
	// Create a roamer instance with our custom YAML decoder
	r := roamer.NewRoamer(
		roamer.WithDecoders(
			NewYAMLDecoder(),
		),
	)

	// HTTP handler that accepts YAML configuration
	http.HandleFunc("/config", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var config Config

		// Parse the YAML request body
		if err := r.Parse(req, &config); err != nil {
			http.Error(w, fmt.Sprintf("Parse error: %v", err), http.StatusBadRequest)
			return
		}

		// Log the parsed configuration
		log.Printf("Received configuration:")
		log.Printf("  Server: %s:%d (TLS: %v)", config.Server.Host, config.Server.Port, config.Server.TLS)
		log.Printf("  Database: %s@%s:%d/%s", config.Database.Username, config.Database.Host,
			config.Database.Port, config.Database.Database)
		log.Printf("  Features: Cache=%v, Metrics=%v, Debug=%v",
			config.Features.EnableCache, config.Features.EnableMetrics, config.Features.DebugMode)

		// Respond with confirmation
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Configuration received successfully!\n\n")
		fmt.Fprintf(w, "Server: %s:%d (TLS: %v)\n", config.Server.Host, config.Server.Port, config.Server.TLS)
		fmt.Fprintf(w, "Database: %s on %s:%d\n", config.Database.Database, config.Database.Host, config.Database.Port)
		fmt.Fprintf(w, "Features: Cache=%v, Metrics=%v, Debug=%v\n",
			config.Features.EnableCache, config.Features.EnableMetrics, config.Features.DebugMode)
	})

	// Example endpoint with usage instructions
	http.HandleFunc("/example", func(w http.ResponseWriter, req *http.Request) {
		examples := []string{
			"Test the custom YAML decoder with curl:",
			"",
			"cat > config.yaml << 'EOF'",
			"server:",
			"  host: localhost",
			"  port: 8080",
			"  tls: true",
			"database:",
			"  driver: postgres",
			"  host: db.example.com",
			"  port: 5432",
			"  database: myapp",
			"  username: admin",
			"  password: secret123",
			"features:",
			"  enable_cache: true",
			"  enable_metrics: true",
			"  debug_mode: false",
			"EOF",
			"",
			"curl -X POST http://localhost:8080/config \\",
			"  -H 'Content-Type: application/yaml' \\",
			"  --data-binary @config.yaml",
			"",
			"Or inline:",
			"",
			"curl -X POST http://localhost:8080/config \\",
			"  -H 'Content-Type: application/yaml' \\",
			"  -d 'server:",
			"  host: localhost",
			"  port: 3000",
			"  tls: false",
			"database:",
			"  driver: mysql",
			"  host: localhost",
			"  port: 3306",
			"  database: testdb",
			"  username: root",
			"  password: pass",
			"features:",
			"  enable_cache: true",
			"  enable_metrics: false",
			"  debug_mode: true'",
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, strings.Join(examples, "\n"))
	})

	// Start the server
	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Printf("Visit http://localhost:8080/example for usage instructions")
	log.Printf("\nThe server accepts YAML configuration at POST /config")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
