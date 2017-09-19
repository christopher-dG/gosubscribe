package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/christopher-dG/gosubscribe"
	irc "github.com/thoj/go-ircevent"
)

var notInitialized = "You're not initialized."

func handlePrivate(e *irc.Event) {
	var msg string
	switch strings.SplitN(e.Message(), " ", 2)[0] {
	case ".sub":
		msg = subscribe(e)
	case ".unsub":
		msg = unsubscribe(e)
	case ".list":
		msg = list(e)
	case ".purge":
		msg = purge(e)
	case ".count":
		msg = count(e)
	case ".top":
		msg = top(e)
	case ".init":
		msg = initUser(e)
	case ".secret":
		msg = getSecret(e)
	case ".register":
		msg = registerUser(e)
	case ".server":
		msg = fmt.Sprintf("[%s %s]", gosubscribe.ServerURL, gosubscribe.ServerURL)
	case ".invite":
		msg = fmt.Sprintf("[%s %s]", gosubscribe.InviteURL, gosubscribe.InviteURL)
	case ".osu":
		msg = fmt.Sprintf("[%s %s]", gosubscribe.OsuUserURL, gosubscribe.OsuUserURL)
	case ".help":
		msg = fmt.Sprintf("[%s %s]", gosubscribe.HelpURL, gosubscribe.HelpURL)
	default:
		return
	}
	bot.Privmsg(e.Nick, msg)
}

// subscribe subscribes the user to the given mappers.
func subscribe(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return notInitialized
	}
	return gosubscribe.Subscribe(user, e.Message(), "")
}

// unsubscribe unsubscribes the user from the given mappers.
func unsubscribe(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return notInitialized
	}
	return gosubscribe.Unsubscribe(user, e.Message(), "")
}

// list displays the mappers that the user is subscribed to.
func list(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return notInitialized
	}
	return gosubscribe.List(user, "")
}

// purge unsubscribes the user from all mappers.
func purge(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return notInitialized
	}
	return gosubscribe.Purge(user, "")
}

// count displays the subscriber counts for the given mappers.
func count(e *irc.Event) string {
	return "Sorry, this command is only available on Discord for now."
	// return gosubscribe.Count(e.Message(), "")
}

// top displays the subscriber counts  for the mappers with the most subscribers.
func top(e *irc.Event) string {
	return "Sorry, this command is only available on Discord for now."
	// return gosubscribe.Top(e.Message())
}

// initUser adds a new user.
func initUser(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err == nil {
		return fmt.Sprintf("You're already initialized; your secret is %s.", user.Secret)
	}
	user, _ = createUser(e.Nick)
	log.Printf(
		".init: initialized new user (osu!): %d -> %s\n",
		user.ID, user.OsuUsername.String,
	)
	return fmt.Sprintf("Initialized; your secret is `%s`.", user.Secret)
}

// getSecret retrieves a user's secret.
func getSecret(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return notInitialized
	}
	secret, err := user.GetSecret()
	if err != nil {
		return "You don't have a secret... this shouldn't happen. Complain to Chris."
	}
	log.Printf(".secret: retrieved secret for %d (length %d)", user.ID, len(user.Secret))
	return fmt.Sprintf("Your secret is: %s.", secret)
}

// registerUser registers a user's osu! username with their existing account.
func registerUser(e *irc.Event) string {
	tokens := strings.Split(e.Message(), " ")
	if len(tokens) == 1 {
		return "You need to supply your secret."
	}
	user, err := gosubscribe.UserFromSecret(tokens[1])
	if err != nil {
		return notInitialized
	}

	if user.OsuUsername.Valid && fmt.Sprint(user.OsuUsername.String) == e.Nick {
		return "You're already registered."
	}
	// This also takes care of name changes.
	user.OsuUsername.String = e.Nick
	user.OsuUsername.Valid = true
	gosubscribe.DB.Save(user)
	log.Printf(
		".register: registered user (osu!): %d -> %s\n",
		user.ID, user.OsuUsername.String,
	)
	return "Registered osu!."
}
