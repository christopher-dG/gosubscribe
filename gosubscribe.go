// gosubscribe is a CRUD library for managing subscriptions.
package gosubscribe

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	DB         *gorm.DB // Connect must be called before using this.
	HelpURL    string   = "https://github.com/christopher-dG/gosubscribe#command-reference"
	ServerURL  string   = "https://discord.gg/qaUhTKJ"
	InviteURL  string   = "https://discordapp.com/oauth2/authorize?client_id=305550679538401280&scope=bot&permissions=3072"
	OsuUserURL string   = "https://osu.ppy.sh/users/3172543"
	osuURL     string   = "https://osu.ppy.sh/api"
	osuKey     string   = os.Getenv("OSU_API_KEY")
)

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

// formatCounts converts a mapper -> subscriber count mapping to an evenly spaced
// table, ordering the counts in descending order.
func FormatCounts(counts map[Mapper]uint) string {
	// First, get the maximum width a mapper's name for formatting.
	maxWidth := -1
	for mapper, _ := range counts {
		if len(mapper.Username) > maxWidth {
			maxWidth = len(mapper.Username)
		}
	}
	maxWidth += 5 // Some padding.

	var s string

	// Now add each line to the output, ordered by count (descending).
	for len(counts) > 0 {
		maxSubs := uint(0)
		var maxSubsMapper Mapper
		for mapper, count := range counts {
			if count >= maxSubs {
				maxSubs = count
				maxSubsMapper = mapper
			}
		}
		padding := strings.Repeat(" ", maxWidth-len(maxSubsMapper.Username))
		var plural string
		if maxSubs != 1 {
			plural = "s"
		}
		s += fmt.Sprintf(
			"%s%s%d subscriber%s\n",
			maxSubsMapper.Username, padding, maxSubs, plural,
		)
		delete(counts, maxSubsMapper)
	}

	return s
}
