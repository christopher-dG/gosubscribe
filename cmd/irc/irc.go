package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/christopher-dG/gosubscribe"
	irc "github.com/thoj/go-ircevent"
)

var bot *irc.Connection

func main() {
	gosubscribe.Connect(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASS"),
	)

	bot = irc.IRC(os.Getenv("IRC_USER"), os.Getenv("IRC_USER"))
	bot.Password = os.Getenv("IRC_PASS")
	err := bot.Connect(fmt.Sprintf("%s:%s", os.Getenv("IRC_SERVER"), os.Getenv("IRC_PORT")))
	if err != nil {
		log.Fatal(err)
	}

	bot.AddCallback("PRIVMSG", handleMessage)
	bot.AddCallback("PING", handlePing)
	bot.Loop()
}

func handleMessage(e *irc.Event) {
	go handlePrivate(e)
}

func handlePing(e *irc.Event) {
	bot.SendRawf("PONG %s", e.Message())
}

// getUser retrieves a user from the database.
func getUser(name string) (*gosubscribe.User, error) {
	user := new(gosubscribe.User)
	gosubscribe.DB.Where("osu_username = ?", name).First(user)
	if user.ID == 0 {
		return user, errors.New("not initialized")
	}
	return user, nil
}

// createUser adds a new user to the database if they don't already exist, and registers
// their osu! username  to their account.
func createUser(name string) (*gosubscribe.User, error) {
	existing, err := getUser(name)
	if err == nil {
		return existing, errors.New("already initialized")
	}
	user := new(gosubscribe.User)
	user.OsuUsername.String = name
	user.OsuUsername.Valid = true
	user.Secret = gosubscribe.GenSecret()
	gosubscribe.DB.Save(user)
	return user, nil
}
