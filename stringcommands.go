package gosubscribe

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

// subscribe subscribes the user to the given mappers.
func Subscribe(user User, body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 {
		if len(mention) > 0 {
			return fmt.Sprintf("%s, you need to supply at least one mapper.", mention)
		} else {
			return "You need to supply at least one mapper."
		}
	}

	names := strings.Split(tokens[1], ",")
	mappers := []Mapper{}
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
		} else {
			return "No mappers were found."
		}
	} else {
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
		} else {
			return fmt.Sprintf("You subscribed to: %s", strings.Join(subscribed, ", "))
		}
	}
}

// unsubscribe unsubscribes the user from the given mappers.
func Unsubscribe(user User, body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 {
		if len(mention) > 0 {
			return fmt.Sprintf("%s, you need to supply at least one mapper.", mention)
		} else {
			return "You need to supply at least one mapper."
		}
	}

	names := strings.Split(tokens[1], ",")
	unsubscribed := []string{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		var mapper Mapper
		DB.Where("lower(username) = lower(?)", name).First(&mapper)
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
				"%s unsubscribed from: %s", mention, strings.Join(unsubscribed, ", "),
			)
		} else {
			return fmt.Sprintf(
				"You unsubscribed from: %s", strings.Join(unsubscribed, ", "),
			)
		}
	} else {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s, you weren't subscribed to any of those mappers.", mention,
			)
		} else {
			return "You weren't subscribed to any of those mappers."
		}
	}
}

// purge unsubscribes the user from all mappers.
func Purge(user User, mention string) string {
	DB.Where("user_id = ?", user.ID).Delete(Subscription{})
	log.Printf(".purge: purged subscriptions for %d\n", user.ID)
	if len(mention) > 0 {
		return fmt.Sprintf("%s is no longer subscribed to any mappers.", mention)
	} else {
		return "You are no longer subscribed to any mappers."
	}
}

// list displays the mappers that the user is subscribed to.
func List(user User, mention string) string {
	mappers := user.ListSubscribed()
	names := []string{}
	for _, mapper := range mappers {
		names = append(names, mapper.Username)
	}

	log.Printf(".list: listing %d subscription(s) for %d\n", len(names), user.ID)
	if len(names) > 0 {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s is subscribed to: %s.", mention, strings.Join(names, ", "),
			)
		}
		return fmt.Sprintf("You're subscribed to: %s", strings.Join(names, ", "))
	} else {
		if len(mention) > 0 {
			return fmt.Sprintf("%s is not subscribed to any mappers.", mention)
		} else {
			return "You're not subscribed to any mappers."
		}
	}
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

// Count displays the subscriber counts for the given mappers.
func Count(body, mention string) string {
	tokens := strings.SplitN(body, " ", 2)
	if len(tokens) == 1 {
		if len(mention) > 0 {
			return fmt.Sprintf(
				"%s, you need to supply at least one mapper.", mention,
			)
		} else {
			return "You need to supply at least one mapper."
		}
	}

	names := strings.Split(tokens[1], ",")
	for i, name := range names {
		names[i] = strings.TrimSpace(name)
	}
	mappers := []Mapper{}
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
			counts[Mapper{Username: name}] = 0
		}
	}

	log.Printf(
		".count: displaying counts for %d mapper(s) (%d found) (from %s)\n",
		len(counts), len(mappers), body,
	)
	return FormatCounts(counts)
}
