package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/christopher-dG/gosubscribe"
)

var testChannel string = os.Getenv("DISCORD_TEST_CHANNEL")
var test bool = os.Getenv("DB_NAME") == "test"
var me string = os.Getenv("DISCORD_ME")

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
