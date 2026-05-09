package main

import (
	"attendance-api/internal/config"
	"attendance-api/internal/database"
	"attendance-api/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

func main() {
	// 1. Initialize Logger
	utils.InitLogger()
	defer utils.Logger.Sync()

	// 2. Initialize Configuration
	cfg := config.LoadConfig()

	// 3. Initialize Database
	database.InitDB(cfg)
	db := database.GetDB()

	fmt.Println("🚀 Iniciando el proceso de sembrado (seeding)...")

	// 4. Determinar la ruta del archivo SQL
	// Intentamos buscarlo en la ruta estándar del proyecto
	seedPath := filepath.Join("database", "seeders", "seed_data.sql")
	
	// Si no existe (ej. estamos ejecutando desde cmd/seed), subimos niveles
	if _, err := os.Stat(seedPath); os.IsNotExist(err) {
		seedPath = filepath.Join("..", "..", "database", "seeders", "seed_data.sql")
	}

	content, err := os.ReadFile(seedPath)
	if err != nil {
		utils.Logger.Fatal("No se pudo leer el archivo de seeder", zap.String("path", seedPath), zap.Error(err))
	}

	// 5. Ejecutar el SQL
	// El archivo puede contener múltiples sentencias, pero SQLX.Exec puede manejar un bloque si el driver lo permite.
	// Para mayor seguridad, podríamos separar por ";" pero los bloques de NOW() y otros podrían ser complejos.
	// Intentaremos ejecutarlo como un script completo.
	
	_, err = db.Exec(string(content))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate key") {
			fmt.Println("⚠️  Parece que algunos datos ya existen en la base de datos.")
			utils.Logger.Warn("Conflicto de duplicidad detectado durante el seeding", zap.Error(err))
		} else {
			utils.Logger.Fatal("Error al ejecutar el seeder", zap.Error(err))
		}
	}

	fmt.Println("✅ Base de datos sembrada correctamente.")
	fmt.Println("🔑 Credenciales por defecto:")
	fmt.Println("   - Usuario: admin@email.com")
	fmt.Println("   - Contraseña: adminpassword")
}
