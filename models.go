package gosubscribe

import "database/sql"

type User struct {
	ID          uint
	DiscordID   sql.NullInt64  `gorm:"unique;index"`
	OsuUsername sql.NullString `gorm:"unique;index"`
	Secret      string         `gorm:"unique"`
	NotifyAll   bool
	MessageOsu  bool
}

type Mapper struct {
	ID       uint
	Username string `gorm:"not null"`
}

type Map struct {
	ID       uint
	MapperID uint // HOW DO FOREIGN KEYS WORK
}

type Subscription struct {
	UserID   uint `gorm:"primary_key"`
	MapperID uint `gorm:"primary_key"`
}
