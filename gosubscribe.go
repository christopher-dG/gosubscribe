// A CRUD library for managing subscriptions.
package gosubscribe

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Connect connects to a given PostreSQL database.
func Connect(host, user, dbname, password string) *gorm.DB {
	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s user=%s dbname=%s password=%s sslmode=disable",
			host, user, dbname, password,
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Mapper{}, &Map{})
	return db
}
