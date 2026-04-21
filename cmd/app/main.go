package main

import (
	"log"
	"prj2/logic"
	"prj2/myHttp"
)

func main() {
	if err := logic.NewEnterprise().Start(); err != nil {
		log.Println("failed to start enterprise:", err)
	}

	enterprise := logic.NewEnterprise()
	handlers := myHttp.NewHTTPHandlers(enterprise)
	server := myHttp.NewHTTTPServer(handlers)
	if err := server.StartServer(); err != nil {
		log.Fatalln(err)
	}
}
