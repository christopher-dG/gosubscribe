package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/christopher-dG/gosubscribe"
	irc "github.com/thoj/go-ircevent"
)

func handlePrivate(e *irc.Event) {
	tokens := strings.SplitN(e.Message(), " ", 2)
	var msg string
	switch tokens[0] {
	case ".sub":
		msg = subscribe(e)
	case ".unsub":
	case ".list":
	case ".purge":
	case ".count":
	case ".top":
	case ".init":
		msg = initUser(e)
	case ".register":
		msg = registerUser(e)
	case ".secret":
		msg = getSecret(e)
	case ".help":
		msg = gosubscribe.HelpURL
	default:
		return
	}
	bot.Privmsg(e.Nick, msg)
}

// subscribe subscribes the user to the given mappers.
func subscribe(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return err.Error()
	}
	tokens := strings.SplitN(e.Message(), " ", 2)
	if len(tokens) == 1 {
		return "You need to supply at least one mapper."
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
		log.Printf(".sub: couldn't find any mappers (from %s)\n", e.Message())
		return fmt.Sprintf("No mappers were found.")
	} else {
		user.Subscribe(mappers)
		subscribed := []string{}
		for _, mapper := range mappers {
			subscribed = append(subscribed, mapper.Username)
		}
		log.Printf(
			".sub: subscribed %d to %d/%d mapper(s) (from %s)\n",
			user.ID, len(mappers), len(names), e.Message(),
		)
		return fmt.Sprintf("You subscribed to: %s.", strings.Join(subscribed, ", "))
	}
}

// unsubscribe unsubscribes the user from the given mappers.
func unsubscribe(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return err.Error()
	}
	tokens := strings.SplitN(e.Message(), " ", 2)
	if len(tokens) == 1 {
		return "You need to supply at least one mapper."
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

	log.Printf(
		".unsub: unsubscribed %d from %d/%d mapper(s) (from %s)\n",
		user.ID, len(unsubscribed), len(names), e.Message(),
	)
	if len(unsubscribed) > 0 {
		return fmt.Sprintf("You unsubscribed from: %s", strings.Join(unsubscribed, ", "))
	} else {
		return "You weren't subscribed to any of those mappers."
	}
}

// purge unsubscribes the user from all mappers.
func purge(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return err.Error()
	}
	gosubscribe.DB.Where("user_id = ?", user.ID).Delete(gosubscribe.Subscription{})
	log.Printf(".purge: purged subscriptions for %d\n", user.ID)
	return "You are no longer subscribed to any mappers."
}

// list displays the mappers that the user is subscribed to.
func list(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return err.Error()
	}
	mappers := user.ListSubscribed()
	names := []string{}
	for _, mapper := range mappers {
		names = append(names, mapper.Username)
	}

	log.Printf(".list: listing %d subscription(s) for %d\n", len(names), user.ID)
	if len(names) > 0 {
		return fmt.Sprintf("You're subscribed to: %s", strings.Join(names, ", "))
	} else {
		return "You're not subscribed to any mappers."
	}
}

// top displays the subscriber counts  for the mappers with the most subscribers.
func top(e *irc.Event) string {
	tokens := strings.SplitN(e.Message(), " ", 2)
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
	log.Printf(".top: displaying top %d mappers (from %s)\n", n, e.Message())
	return gosubscribe.FormatCounts(gosubscribe.Top(n))
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
		if !gosubscribe.HasMapper(counts, name) {
			counts[gosubscribe.Mapper{Username: name}] = 0
		}
	}

	log.Printf(
		".count: displaying counts for %d mapper(s) (%d found) (from %s)\n",
		len(counts), len(mappers), m.Content,
	)
	return gosubscribe.FormatCounts(counts)
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

// registerUser registers a user's osu! username with their existing account.
func registerUser(e *irc.Event) string {
	tokens := strings.Split(e.Message(), " ")
	if len(tokens) == 1 {
		return "You need to supply your secret."
	}
	user, err := gosubscribe.UserFromSecret(tokens[1])
	if err != nil {
		return err.Error()
	}

	if user.OsuUsername.Valid && fmt.Sprint(user.OsuUsername.String) == e.Nick {
		return "You're already registered."
	}
	// This also takes care of name changes.
	user.OsuUsername.String = e.Nick
	user.OsuUsername.Valid = true
	gosubscribe.DB.Save(&user)
	log.Printf(
		".register: registered user (osu!): %d -> %s\n",
		user.ID, user.OsuUsername.String,
	)
	return "Registered osu!."
}

// getSecret retrieves a user's secret.
func getSecret(e *irc.Event) string {
	user, err := getUser(e.Nick)
	if err != nil {
		return err.Error()
	}
	secret, err := gosubscribe.GetSecret(user)
	if err != nil {
		return err.Error()
	}
	log.Printf(".secret: retrieved secret for %d (length %d)", user.ID, len(user.Secret))
	return fmt.Sprintf("Your secret is: `%s`.", secret)
}
