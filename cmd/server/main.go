package main

import (
	"fmt"
	"log"

	"github.com/sukryu/IV-auth-services/configs"
)

func main() {
	// Config is loaded via Viper in configs package init
	cfg := configs.GlobalConfig
	fmt.Printf("Loaded config: DB Host=%s, Redis Addr=%s\n", cfg.DB.Host, cfg.Redis.Addr)
	log.Println("Authentication Service starting...")
}
