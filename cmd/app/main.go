package main

import (
	"log"
	"prj2/logic"
)

func main() {
	if err := logic.NewEnterprise().Start(); err != nil {
		log.Println("failed to start enterprise:", err)
	}
}
