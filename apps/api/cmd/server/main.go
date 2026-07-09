package main

import (
	"log"

	"github.com/fulltank-garage/linora/apps/api/internal/analysis"
	"github.com/fulltank-garage/linora/apps/api/internal/httpapi"
)

func main() {
	router := httpapi.NewRouter(analysis.NewService())
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
