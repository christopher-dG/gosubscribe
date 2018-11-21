package gosubscribe

import (
	"errors"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Only needed for the driver.
)

// DB (the database instance) cannot be used before Connect is called.
var DB *gorm.DB

// Connect connects to a given PostreSQL database.
// PGxxx environment variables must be set.
func Connect(host, user, dbname, password string) {
	db, err := gorm.Open("postgres", "postgresql://")
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Mapper{}, &Mapset{}, &Subscription{})
	DB = db // From now on, we can access the database from anywhere via DB.
}

// GetUser retrieves a user by their ID.
func GetUser(id uint) (*User, error) {
	user := new(User)
	DB.Where("ID = ?", id).First(user)
	if user.ID == 0 {
		return nil, errors.New("no user found")
	}
	return user, nil
}

// Subscribe subscribes a user to a list of mappers.
func (user *User) Subscribe(mappers []*Mapper) {
	for _, mapper := range mappers {
		DB.Table("subscriptions").Create(&Subscription{user.ID, mapper.ID})
	}
}

// Unsubscribe unsubscribes a user from a list of mappers.
func (user *User) Unsubscribe(mappers []*Mapper) {
	for _, mapper := range mappers {
		DB.Delete(&Subscription{user.ID, mapper.ID})
	}
}

// ListSubscribed gets all mappers that a user is subscribed to.
func (user *User) ListSubscribed() []*Mapper {
	var mappers []*Mapper
	DB.Table("subscriptions s").Joins("INNER JOIN mappers m ON s.mapper_id = m.id").
		Select("m.id, m.username").Where("s.user_id = ?", user.ID).Find(&mappers)
	return mappers
}

// Purge unsubscribes a user from all mappers.
func (user *User) Purge() {
	DB.Table("subscriptions").Where("user_id = ?", user.ID).Delete(Subscription{})
}

// Count gets the number of users that are subscribed to a mapper.
func (mapper *Mapper) Count() uint {
	var count uint
	DB.Model(&Subscription{}).Where("mapper_id = ?", mapper.ID).Count(&count)
	return count
}

// GetCounts gets the subscriber counts for a list of mappers.
func GetCounts(mappers []*Mapper) map[*Mapper]uint {
	counts := make(map[*Mapper]uint)
	for _, mapper := range mappers {
		counts[mapper] = mapper.Count()
	}
	return counts
}

// TopCounts gets the n mappers with the most subscribers and their subscription counts.
func TopCounts(n int) map[*Mapper]uint {
	counts := make(map[*Mapper]uint)
	var subs []Subscription
	// TODO: Figure out how to properly build this query.
	DB.Raw(fmt.Sprintf("SELECT mapper_id, COUNT(*) FROM subscriptions GROUP BY mapper_id ORDER BY COUNT DESC LIMIT %d", n)).Find(&subs)
	for _, sub := range subs {
		mapper := new(Mapper)
		DB.Where("id = ?", sub.MapperID).First(mapper)
		counts[mapper] = mapper.Count()
	}
	return counts
}

// SetNotifyAll sets the users preference for receiving notifications for map updates that
// are not new uploads or ranked status changes.
func (user *User) SetNotifyAll(pref bool) {
	user.NotifyAll = pref
	DB.Save(user)
}

// SetMessageOsu sets the user's preference for receiving messages via osu! or via Discord.
func (user *User) SetMessageOsu(pref bool) {
	user.MessageOsu = pref
	DB.Save(user)
}
