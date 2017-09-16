package main

import (
	"fmt"
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
	switch strings.Split(m.Message.Content, " ")[0] {
	case ".init":
		msg = initUser(m)
	case ".register":
		msg = registerUser(m)
	case ".secret":
		msg = getSecret(m)
	case ".help":
		msg = privateHelp(m)
	default:
		msg = "I don't recognize that command, try `.help`."
	}
	s.ChannelMessageSend(m.ChannelID, msg)
}

// privateHelp returns a help message for authentication.
func privateHelp(m *discordgo.MessageCreate) string {
	tokens := strings.Split(m.Message.Content, " ")
	if len(tokens) == 1 {
		return "Commands: `.init`, `.register`, `.secret`.\nTry `.help auth` for details."
	}
	if tokens[1] == "auth" {
		return authHelp
	} else {
		return "Unrecognized argument; try `.help`."
	}
}

// initUser adds a new user.
func initUser(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err == nil {
		return fmt.Sprintf("You're already initialized; your secret is `%s`.", user.Secret)
	}
	user, _ = createUser(m.Author)
	return fmt.Sprintf("Initialized; your secret is `%s`.", user.Secret)
}

// registerUser registers a user's Discord ID with their existing account.
func registerUser(m *discordgo.MessageCreate) string {
	tokens := strings.Split(m.Message.Content, " ")
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
