package main

import (
	"attendance-api/internal/config"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.LoadConfig()
	
	m, err := migrate.New(
		"file://database/migrations",
		cfg.DBURL,
	)
	if err != nil {
		fmt.Printf("❌ Error al crear instancia: %v\n", err)
		return
	}

	version := 10101000005 // La última exitosa según la inspección
	fmt.Printf("🔧 Forzando versión de migración a %d...\n", version)
	
	err = m.Force(version)
	if err != nil {
		fmt.Printf("❌ Error al forzar: %v\n", err)
		return
	}

	fmt.Println("✅ Versión forzada correctamente. Ahora ejecuta el seed.")
}
