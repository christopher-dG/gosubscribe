package gosubscribe

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
)

var (
	// HelpURL is a link to the command reference.
	HelpURL = "https://github.com/christopher-dG/gosubscribe#command-reference"
	// ServerURL is an invite to the Discord server.
	ServerURL = "https://discord.gg/qaUhTKJ"
	// InviteURL is a link to invite subscription-bot to another server.
	InviteURL = "https://discordapp.com/oauth2/authorize?client_id=305550679538401280&scope=bot&permissions=3072"
	// OsuUserURL is a link to the bot's userpage.
	OsuUserURL = "https://osu.ppy.sh/users/3172543"
	noMappers  = "You need to supply at least one mapper."
)

// Subscribe subscribes the user to the given mappers.
func Subscribe(user *User, body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 {
		if len(mention) > 0 {
			return fmt.Sprintf("%s, you need to supply at least one mapper.", mention)
		}
		return noMappers
	}

	names := GetTokens(tokens[1])
	mappers := []*Mapper{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		mapper, err := GetMapper(name)
		if err == nil {
			mappers = append(mappers, mapper)
		} else {
			log.Println(err)
		}
	}

	if len(mappers) == 0 {
		log.Printf(".sub: couldn't find any mappers (from %s)\n", body)
		if len(mention) > 0 {
			return fmt.Sprintf("%s, no mappers were found.", mention)
		}
		return "No mappers were found."
	}
	user.Subscribe(mappers)
	subscribed := []string{}
	for _, mapper := range mappers {
		subscribed = append(subscribed, mapper.Username)
	}
	log.Printf(
		".sub: subscribed %d to %d/%d mapper(s) (from %s)\n",
		user.ID, len(mappers), len(names), body,
	)
	if len(mention) > 0 {
		return fmt.Sprintf(
			"%s subscribed to: %s.", mention, strings.Join(subscribed, ", "),
		)
	}
	return fmt.Sprintf("You subscribed to: %s.", strings.Join(subscribed, ", "))
}

// Unsubscribe unsubscribes the user from the given mappers.
func Unsubscribe(user *User, body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 {
		if len(mention) > 0 {
			return fmt.Sprintf("%s, you need to supply at least one mapper.", mention)
		}
		return noMappers
	}

	names := GetTokens(tokens[1])
	unsubscribed := []string{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		mapper := new(Mapper)
		DB.Where("lower(username) = lower(?)", name).First(mapper)
		if mapper.ID != 0 {
			unsubscribed = append(unsubscribed, mapper.Username)
			DB.Delete(Subscription{user.ID, mapper.ID})
		}
	}

	log.Printf(
		".unsub: unsubscribed %d from %d/%d mapper(s) (from %s)\n",
		user.ID, len(unsubscribed), len(names), body,
	)
	if len(unsubscribed) > 0 {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s unsubscribed from: %s.", mention, strings.Join(unsubscribed, ", "),
			)
		}
		return fmt.Sprintf(
			"You unsubscribed from: %s.", strings.Join(unsubscribed, ", "),
		)
	}
	if len(mention) > 0 {
		return fmt.Sprintf(
			"%s, you weren't subscribed to any of those mappers.", mention,
		)
	}
	return "You weren't subscribed to any of those mappers."
}

// List displays the mappers that the user is subscribed to.
func List(user *User, mention string) string {
	mappers := user.ListSubscribed()
	names := []string{}
	for _, mapper := range mappers {
		names = append(names, mapper.Username)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })

	log.Printf(".list: listing %d subscription(s) for %d\n", len(names), user.ID)
	if len(names) > 0 {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s is subscribed to: %s.", mention, strings.Join(names, ", "),
			)
		}
		return fmt.Sprintf("You're subscribed to: %s.", strings.Join(names, ", "))
	}
	if len(mention) > 0 {
		return fmt.Sprintf("%s is not subscribed to any mappers.", mention)
	}
	return "You're not subscribed to any mappers."
}

// Purge unsubscribes the user from all mappers.
func Purge(user *User, mention string) string {
	DB.Where("user_id = ?", user.ID).Delete(Subscription{})
	log.Printf(".purge: purged subscriptions for %d\n", user.ID)
	if len(mention) > 0 {
		return fmt.Sprintf("%s is no longer subscribed to any mappers.", mention)
	}
	return "You are no longer subscribed to any mappers."
}

// Count displays the subscriber counts for the given mappers.
func Count(body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s, you need to supply at least one mapper.", mention,
			)
		}
		return noMappers
	}

	names := GetTokens(tokens[1])
	for i, name := range names {
		names[i] = strings.TrimSpace(name)
	}
	mappers := []*Mapper{}
	for _, name := range names {
		mapper, err := MapperFromDB(name)
		if err == nil {
			mappers = append(mappers, mapper)
		} else {
			log.Println(err)
		}
	}

	counts := GetCounts(mappers)
	// Add 0 counts for each mapper not found in the DB.
	for _, name := range names {
		if !HasMapper(counts, name) {
			counts[&Mapper{Username: name}] = 0
		}
	}

	log.Printf(
		".count: displaying counts for %d mapper(s) (%d found) (from %s)\n",
		len(counts), len(mappers), body,
	)
	return FormatCounts(counts)
}

// Top displays the subscriber counts  for the mappers with the most subscribers.
func Top(body string) string {
	tokens := strings.SplitN(body, " ", 2)
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
	log.Printf(".top: displaying top %d mappers (from %s)\n", n, body)
	return FormatCounts(TopCounts(n))
}

// NotificationPreference sets the users preference for receiving notifications for map
// updates that are not new uploads or ranked status changes.
func NotificationPreference(user *User, body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 ||
		(!strings.EqualFold(tokens[1], "n") && !strings.EqualFold(tokens[1], "y")) {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s, you need to supply a preference ('y' or 'n').", mention,
			)
		}
		return "You need to supply a preference ('y' or 'n')."
	}
	// Now we know that the argument is a valid option.
	pref := strings.EqualFold(tokens[1], "y")
	var prefString string
	if pref {
		prefString = "all map updates"
	} else {
		prefString = "new uploads and ranked status updates"
	}
	user.SetNotifyAll(pref)
	if len(mention) > 0 {
		return fmt.Sprintf(
			"%s, you will receive notifications for: %s.", mention, prefString,
		)
	}
	return fmt.Sprintf("You will receive notifications for %s.", prefString)
}

// NotificationPlatform sets the user's platform preference for receiving notifications.
func NotificationPlatform(user *User, body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 || (!strings.EqualFold(tokens[1], "discord") &&
		!strings.EqualFold(tokens[1], "osu!")) {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s, you need to supply a preference ('discord' or 'osu!').", mention,
			)
		}
		return "You need to supply a preference ('discord' or 'osu!')."
	}
	// Now we know that the argument is a valid option.
	pref := strings.EqualFold(tokens[1], "osu!")
	var prefString string
	if pref {
		prefString = "osu!"
	} else {
		prefString = "Discord"
	}
	user.SetMessageOsu(pref)
	if len(mention) > 0 {
		return fmt.Sprintf(
			"%s, you set your notification platform to: %s.", mention, prefString,
		)
	}
	return fmt.Sprintf("You set your notification platform to: %s.", prefString)
}
