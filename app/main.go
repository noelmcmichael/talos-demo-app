package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "dev"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Talos Demo App</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 800px;
            margin: 100px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            text-align: center;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 20px;
            padding: 60px;
            backdrop-filter: blur(10px);
        }
        h1 { font-size: 48px; margin: 0 0 20px 0; }
        .version { font-size: 24px; opacity: 0.9; }
        .timestamp { font-size: 16px; opacity: 0.7; margin-top: 20px; }
        .status { color: #4ade80; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Talos Demo App</h1>
        <div class="version">Version: %s</div>
        <div class="timestamp">Deployed: %s</div>
        <div class="status">✅ BGP LoadBalancer Working</div>
    </div>
</body>
</html>`, version, time.Now().Format(time.RFC1123))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	port := "8080"
	log.Printf("Starting server on port %s (version: %s)", port, version)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
// CI/CD test trigger
