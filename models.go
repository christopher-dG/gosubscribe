package gosubscribe

import "database/sql"

// TODO: Proper foreign keys.

type User struct {
	ID          uint
	DiscordID   sql.NullInt64  `gorm:"unique;index"`
	OsuUsername sql.NullString `gorm:"unique;index"`
	Secret      string         `gorm:"unique"`
	NotifyAll   bool
	MessageOsu  bool
}

type Mapper struct {
	ID       uint   `json:"user_id,string"`
	Username string `gorm:"not null" json:"username"`
}

type Map struct {
	ID       uint `json:"beatmapset_id,string"`
	MapperID uint `json:"-"` // Need to fill this field manually.
	Status   int  `json:"approved,string"`
}

type Subscription struct {
	UserID   uint `gorm:"primary_key"`
	MapperID uint `gorm:"primary_key"`
}
