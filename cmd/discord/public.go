package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
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
	case ".init":
		fallthrough
	case ".register":
		fallthrough
	case ".secret":
		msg = fmt.Sprintf(
			"%s, this command belongs in a private message.",
			m.Author.Mention(),
		)
	case ".help":
		msg = gosubscribe.HelpURL
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

// top displays the subscriber counts  for the mappers with the most subscribers.
func top(m *discordgo.MessageCreate) string {
	tokens := strings.SplitN(m.Content, " ", 2)
	var n int
	if len(tokens) == 1 {
		n = 5
	} else {
		parsed, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil || parsed <= 0 {
			n = 5
		} else {
			n = int(math.Min(float64(parsed), 25))
		}
	}
	return formatCounts(gosubscribe.Top(n))
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

// formatCounts converts a mapper -> subscriber count mapping to a fenced code block,
// ordering the counts in descending order.
func formatCounts(counts map[gosubscribe.Mapper]uint) string {
	// First, get the maximum width a mapper's name for formatting.
	maxWidth := -1
	for mapper, _ := range counts {
		if len(mapper.Username) > maxWidth {
			maxWidth = len(mapper.Username)
		}
	}
	maxWidth += 5 // Some padding.

	s := "```\n"

	// Now add each line to the output, ordered by count (descending).
	for len(counts) > 0 {
		maxSubs := uint(0)
		var maxSubsMapper gosubscribe.Mapper
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

	return s + "```"
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
