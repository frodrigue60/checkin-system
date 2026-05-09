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

	fmt.Println("🧹 Limpiando tabla de migraciones...")
	_, err := db.Exec("DROP TABLE IF EXISTS schema_migrations")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	fmt.Println("✅ Tabla schema_migrations eliminada. Intenta ejecutar el seed de nuevo.")
}
