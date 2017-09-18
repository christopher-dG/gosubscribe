package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/christopher-dG/gosubscribe"
)

// handlePrivate responds to messages sent via private channels (DMs).
func handlePrivate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if test && m.Author.ID != me {
		return
	}
	var msg string
	switch strings.Split(m.Content, " ")[0] {
	case ".init":
		msg = initUser(m)
	case ".register":
		msg = registerUser(m)
	case ".secret":
		msg = getSecret(m)
	case ".help":
		msg = gosubscribe.HelpURL
	default:
		msg = "Unrecognized command, maybe it belongs in a public channel. Try `.help`."
	}
	s.ChannelMessageSend(m.ChannelID, msg)
}

// initUser adds a new user.
func initUser(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err == nil {
		return fmt.Sprintf("You're already initialized; your secret is `%s`.", user.Secret)
	}
	user, _ = createUser(m.Author)
	log.Printf("initialized new user: %d -> %d\n", user.ID, user.DiscordID)
	return fmt.Sprintf("Initialized; your secret is `%s`.", user.Secret)
}

// registerUser registers a user's Discord ID with their existing account.
func registerUser(m *discordgo.MessageCreate) string {
	tokens := strings.Split(m.Content, " ")
	if len(tokens) == 1 {
		return "You need to supply your secret."
	}
	user, err := gosubscribe.UserFromSecret(tokens[1])
	if err != nil {
		return err.Error()
	}

	if user.DiscordID.Valid && fmt.Sprint(user.DiscordID.Int64) == m.Author.ID {
		return "You're already registered."
	}
	id, _ := strconv.ParseInt(m.Author.ID, 10, 64)
	user.DiscordID.Int64 = id
	user.DiscordID.Valid = true
	gosubscribe.DB.Save(&user)
	log.Printf("registered user: %d -> %d\n", user.ID, user.DiscordID)
	return "Registered Discord."
}

// getSecret retrieves a user's secret.
func getSecret(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return err.Error()
	}
	secret, err := gosubscribe.GetSecret(user)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("Your secret is: `%s`.", secret)
}
