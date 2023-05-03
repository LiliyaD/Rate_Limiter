package main

import (
	"log"

	"github.com/LiliyaD/Rate_Limiter/config"
	"github.com/LiliyaD/Rate_Limiter/internal/app"
)

func main() {
	log.Println("Starting service")

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(app.NewApp(cfg).Run())
}
