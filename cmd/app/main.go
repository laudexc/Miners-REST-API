package main

import (
	"log"
	"prj2/logic"
	"prj2/myHttp"
)

func main() {
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
