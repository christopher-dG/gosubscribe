package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/christopher-dG/gosubscribe"
)

var (
	testChannel string = os.Getenv("DISCORD_TEST_CHANNEL")
	test        bool   = os.Getenv("DB_NAME") == "test"
	me          string = os.Getenv("DISCORD_ME")
	authHelp    string = "https://github.com/christopher-dG/gosubscribe#authentication"
)

func main() {
	gosubscribe.Connect(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASS"),
	)

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = discord.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// messageCreate handles incoming messages.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if isPrivate(s, m) {
		handlePrivate(s, m)
	} else {
		handlePublic(s, m)
	}
}

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

// handlePublic responds to messages sent via public channels.
func handlePublic(s *discordgo.Session, m *discordgo.MessageCreate) {
	if test && m.ChannelID != testChannel {
		return
	}
}

// isPrivate determines whether or not an incoming message is in a private channel.
func isPrivate(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return true // TBD.
	}
	return channel.Type == discordgo.ChannelTypeDM
}

// getUser retrieves a user from the database.
func getUser(dUser *discordgo.User) (gosubscribe.User, error) {
	var (
		err  error
		user gosubscribe.User
	)
	gosubscribe.DB.Where("discord_id = ?", dUser.ID).Find(&user)
	if user.ID == 0 {
		return user, errors.New("You aren't initialized. Try `.help auth`.")
	}
	return user, err
}

// createUser adds a new user to the database if they don't already exist, and registers
// their Discord ID to their account.
func createUser(dUser *discordgo.User) (gosubscribe.User, error) {
	existing, err := getUser(dUser)
	if err == nil {
		return existing, errors.New("Already initialized.")
	}
	id, _ := strconv.ParseInt(dUser.ID, 10, 64)
	user := new(gosubscribe.User)
	user.DiscordID.Int64 = id
	user.DiscordID.Valid = true
	user.Secret = gosubscribe.GenSecret()
	gosubscribe.DB.Save(&user)
	return *user, nil

}
