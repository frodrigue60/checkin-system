package main

import (
	"attendance-api/internal/config"
	"fmt"
	"os"
	"strconv"

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

	version := 20260509160500
	if len(os.Args) > 1 {
		parsed, err := strconv.Atoi(os.Args[1])
		if err == nil {
			version = parsed
		} else {
			fmt.Printf("⚠️ No se pudo parsear '%s' como versión numérica, usando defecto %d\n", os.Args[1], version)
		}
	}
	
	fmt.Printf("🔧 Forzando versión de migración a %d...\n", version)
	
	err = m.Force(version)
	if err != nil {
		fmt.Printf("❌ Error al forzar: %v\n", err)
		return
	}

	fmt.Println("✅ Versión forzada correctamente.")
}
