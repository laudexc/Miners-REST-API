package main

import (
	"log"
	"net/http"
	"os"
	"prj2/logic"
	"prj2/myHttp"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

	enterprise := logic.NewEnterprise()
	if err := enterprise.Start(); err != nil {
		log.Println("failed to start enterprise:", err)
	}

	handlers := myHttp.NewHTTPHandlers(enterprise)
	server := myHttp.NewHTTTPServer(handlers)
	if err := server.StartServer(); err != nil {
		log.Fatalln(err)
	}
}

func runHealthcheck() {
	client := http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:8080/health")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
