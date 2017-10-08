package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/christopher-dG/gosubscribe"
)

// handlePublic responds to messages sent via public channels.
func handlePublic(s *discordgo.Session, m *discordgo.MessageCreate) {
	if test && m.ChannelID != testChannel {
		return
	}
	var msg string
	switch strings.Split(m.Message.Content, " ")[0] {
	case ".sub":
		msg = subscribe(m)
	case ".unsub":
		msg = unsubscribe(m)
	case ".list":
		msg = list(m)
	case ".purge":
		msg = purge(m)
	case ".count":
		msg = count(m)
	case ".top":
		msg = top(m)
	case ".notifyall":
		msg = notifyAll(m)
	case ".message":
		msg = message(m)
	case ".server":
		msg = gosubscribe.ServerURL
	case ".invite":
		msg = fmt.Sprintf("<%s>", gosubscribe.InviteURL)
	case ".osu":
		msg = gosubscribe.OsuUserURL
	case ".help":
		msg = fmt.Sprintf("<%s>", gosubscribe.HelpURL)
	case ".init":
		fallthrough
	case ".secret":
		fallthrough
	case ".register":
		msg = fmt.Sprintf(
			"%s, this command belongs in a private message.", m.Author.Mention(),
		)
	}
	s.ChannelMessageSend(m.ChannelID, msg)
}

// subscribe subscribes the user to the given mappers.
func subscribe(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return fmt.Sprintf("%s, you're not initialized.", m.Author.Mention())
	}
	return escape(gosubscribe.Subscribe(user, m.Content, m.Author.Mention()))
}

// unsubscribe unsubscribes the user from the given mappers.
func unsubscribe(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return fmt.Sprintf("%s, you're not initialized.", m.Author.Mention())
	}
	return escape(gosubscribe.Unsubscribe(user, m.Content, m.Author.Mention()))
}

// list displays the mappers that the user is subscribed to.
func list(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return fmt.Sprintf("%s, you're not initialized.", m.Author.Mention())
	}
	return escape(gosubscribe.List(user, m.Author.Mention()))
}

// purge unsubscribes the user from all mappers.
func purge(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return fmt.Sprintf("%s, you're not initialized.", m.Author.Mention())
	}
	return gosubscribe.Purge(user, m.Author.Mention())
}

// count displays the subscriber counts for the given mappers.
func count(m *discordgo.MessageCreate) string {
	res := gosubscribe.Count(m.Content, m.Author.Mention())
	// Only fence the result if it's actually a table.
	if strings.HasPrefix(res, m.Author.Mention()) {
		return res
	}
	return fmt.Sprintf("```%s```", res)
}

// top displays the subscriber counts  for the mappers with the most subscribers.
func top(m *discordgo.MessageCreate) string {
	return fmt.Sprintf("```%s```", gosubscribe.Top(m.Content))
}

// notifyAll sets the user's notification type preference.
func notifyAll(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return fmt.Sprintf("%s, you're not initialized.", m.Author.Mention())
	}
	return gosubscribe.NotificationPreference(user, m.Content, m.Author.Mention())
}

// message sets the user's notification platform preference.
func message(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return fmt.Sprintf("%s, you're not initialized.", m.Author.Mention())
	}
	return gosubscribe.NotificationPlatform(user, m.Content, m.Author.Mention())
}
