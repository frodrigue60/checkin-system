package main

import (
	"attendance-api/internal/config"
	"attendance-api/internal/database"
	"fmt"
)

func main() {
	cfg := config.LoadConfig()
	database.InitDB(cfg)
	db := database.GetDB()

	fmt.Println("📋 Listando tablas existentes...")
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("- %s\n", name)
	}
}
