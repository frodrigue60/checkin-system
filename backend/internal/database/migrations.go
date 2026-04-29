package database

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations proporciona una utilidad para ejecutar las migraciones.
// NOTA: Esta función no se llama automáticamente al inicio por petición del usuario,
// ya que la estructura actual de la DB ya coincide con estas migraciones.
func RunMigrations(dbURL string) {
	log.Println("Iniciando verificación de migraciones...")

	m, err := migrate.New(
		"file://database/migrations",
		dbURL,
	)
	if err != nil {
		log.Printf("Error al crear instancia de migración: %v", err)
		return
	}

	// m.Up() aplicaría las migraciones pendientes. 
	// Se deja comentado o disponible para uso manual.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Error al ejecutar migraciones Up: %v", err)
	} else {
		log.Println("Migraciones verificadas/aplicadas correctamente")
	}
	
	log.Println("Utilidad de migraciones lista. (Ejecución automática activada)")
}
