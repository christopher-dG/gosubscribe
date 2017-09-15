package gosubscribe

import "database/sql"

type User struct {
	ID              uint
	DiscordID       sql.NullInt64 `gorm:"unique;index"`
	DiscordDisc     sql.NullInt64
	DiscordUsername sql.NullString
	OsuID           sql.NullInt64  `gorm:"unique"`
	OsuUsername     sql.NullString `gorm:"unique;index"`
	NotifyAll       bool           `sql:"default:false"`
	Subscriptions   []Mapper       `gorm:"many2many:subscriptions"`
}

type Mapper struct {
	ID       uint
	Username string `gorm:"not null"`
}

type Map struct {
	ID       uint
	MapperID uint // HOW DO FOREIGN KEYS WORK
}
