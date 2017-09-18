// gosubscribe is a CRUD library for managing subscriptions.
package gosubscribe

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB // Connect must be called before using this.
var HelpURL string = "<https://github.com/christopher-dG/gosubscribe#command-reference>"

// Connect connects to a given PostreSQL database.
func Connect(host, user, dbname, password string) {
	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s user=%s dbname=%s sslmode=disable password=%s",
			host, user, dbname, password,
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Mapper{}, &Map{}, &Subscription{})
	DB = db // From now on, we can access the database from anywhere via DB.
}
