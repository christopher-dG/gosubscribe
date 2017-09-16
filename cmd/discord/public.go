package main

import (
	"fmt"
	"log"
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
	case ".help":
		msg = publicHelp(m)
	}

	s.ChannelMessageSend(m.ChannelID, msg)
}

// subscribe subscribes the user to the given mappers.
func subscribe(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return err.Error()
	}
	tokens := strings.SplitN(m.Content, " ", 2)
	if len(tokens) == 1 {
		return fmt.Sprintf(
			"%s, you need to supply at least one mapper.",
			m.Author.Mention(),
		)
	}

	names := strings.Split(tokens[1], ",")
	mappers := []gosubscribe.Mapper{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		mapper, err := gosubscribe.GetMapper(name)
		if err == nil {
			mappers = append(mappers, mapper)
		} else {
			log.Println(err)
		}
	}

	if len(mappers) == 0 {
		return fmt.Sprintf("%s, no mappers were found.", m.Author.Mention())
	} else {
		user.Subscribe(mappers)
		subscribed := []string{}
		for _, mapper := range mappers {
			subscribed = append(subscribed, mapper.Username)
		}
		return fmt.Sprintf(
			"%s subscribed to: %s.",
			m.Author.Mention(), strings.Join(subscribed, ", "),
		)
	}
}

// unsubscribe unsubscribes the user from the given mappers.
func unsubscribe(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return err.Error()
	}
	tokens := strings.SplitN(m.Content, " ", 2)
	if len(tokens) == 1 {
		return fmt.Sprintf(
			"%s, you need to supply at least one mapper.",
			m.Author.Mention(),
		)
	}

	names := strings.Split(tokens[1], ",")
	unsubscribed := []string{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		var mapper gosubscribe.Mapper
		gosubscribe.DB.Where("lower(username) = lower(?)", name).First(&mapper)
		if mapper.ID != 0 {
			unsubscribed = append(unsubscribed, mapper.Username)
			gosubscribe.DB.Delete(gosubscribe.Subscription{user.ID, mapper.ID})
		}

	}

	if len(unsubscribed) > 0 {
		return fmt.Sprintf(
			"%s unsubscribed from: %s",
			m.Author.Mention(), strings.Join(unsubscribed, ", "),
		)
	} else {
		return fmt.Sprintf(
			"%s, you weren't subscribed to any of those mappers.", m.Author.Mention(),
		)
	}
}

// purge unsubscribes the user from all mappers.
func purge(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return err.Error()
	}
	gosubscribe.DB.Where("user_id = ?", user.ID).Delete(gosubscribe.Subscription{})
	return fmt.Sprintf("%s is no longer subscribed to any mappers.", m.Author.Mention())
}

// list displays the mappers that the user is subscribed to.
func list(m *discordgo.MessageCreate) string {
	user, err := getUser(m.Author)
	if err != nil {
		return err.Error()
	}
	mappers := user.ListSubscribed()
	names := []string{}
	for _, mapper := range mappers {
		names = append(names, mapper.Username)
	}
	if len(names) > 0 {
		return fmt.Sprintf(
			"%s is subscribed to: %s",
			m.Author.Mention(), strings.Join(names, ", "),
		)
	} else {
		return fmt.Sprintf("%s is not subscribed to any mappers.", m.Author.Mention())
	}
}

// count displays the subscriber counts for the given mappers.
func count(m *discordgo.MessageCreate) string {
	tokens := strings.SplitN(m.Content, " ", 2)
	if len(tokens) == 1 {
		return fmt.Sprintf(
			"%s, you need to supply at least one mapper.",
			m.Author.Mention(),
		)
	}

	names := strings.Split(tokens[1], ",")
	for i, name := range names {
		names[i] = strings.TrimSpace(name)
	}
	mappers := []gosubscribe.Mapper{}
	for _, name := range names {
		mapper, err := gosubscribe.MapperFromDB(name)
		if err == nil {
			mappers = append(mappers, mapper)
		} else {
			log.Println(err)
		}
	}

	counts := gosubscribe.Count(mappers)
	// Add 0 counts for each mapper not found in the DB.
	for _, name := range names {
		if !hasMapper(counts, name) {
			counts[gosubscribe.Mapper{Username: name}] = 0
		}
	}
	return formatCounts(counts)
}

// formatCounts converts a mapper -> subscriber count mapping to a fenced code block.
func formatCounts(counts map[gosubscribe.Mapper]uint) string {
	// TODO: Sort in descending order.
	maxWidth := 0
	for mapper, _ := range counts {
		if len(mapper.Username) > maxWidth {
			maxWidth = len(mapper.Username)
		}
	}
	maxWidth += 5
	s := "```\n"
	for mapper, count := range counts {
		var plural string
		if count != 1 {
			plural = "s"
		}
		padding := strings.Repeat(" ", maxWidth-len(mapper.Username))
		s += fmt.Sprintf("%s%s%d subscriber%s\n", mapper.Username, padding, count, plural)
	}
	return s + "```"
}

// publicHelp dislays a help message for commands.
func publicHelp(m *discordgo.MessageCreate) string {
	return ""
}

// hasMapper determines whether or not the map contains a mapper key with the given name.
func hasMapper(mappers map[gosubscribe.Mapper]uint, name string) bool {
	for mapper, _ := range mappers {
		if strings.EqualFold(mapper.Username, name) {
			return true
		}
	}
	return false
}
