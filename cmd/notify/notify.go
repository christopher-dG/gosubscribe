package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/christopher-dG/gosubscribe"
)

var (
	statusMap = map[int]string{
		-2: "Graveyard",
		-1: "WIP",
		0:  "Pending",
		1:  "Ranked",
		2:  "Approved",
		3:  "Qualified",
		4:  "Loved",
	}
	logChannel    = os.Getenv("DISCORD_LOG_CHANNEL")
	searchKey     = os.Getenv("OSUSEARCH_API_KEY")
	searchURL     = "http://osusearch.com/api/search"
	notifications = make(map[string]map[uint][]*OsuSearchMapset)
	today         = time.Now().Format("2006-01-02")
	bot           *discordgo.Session
)

// SearchResult is the JSON structure returned by osusearch.com.
type SearchResult struct {
	Mapsets []*OsuSearchMapset `json:"beatmaps"`
}

// OsuSearchMapset is an osu! beatmapset that comes from osusearch.com
// (some JSON fields are different).
type OsuSearchMapset struct {
	ID     uint   `json:"beatmapset_id"`
	Mapper string `json:"mapper"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Status int    `json:"beatmap_status"`
}

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
	} else {
		bot = discord
	}

	logMsg(fmt.Sprintf("Starting run for %s.", today))

	notifications["new"] = make(map[uint][]*OsuSearchMapset)
	notifications["status"] = make(map[uint][]*OsuSearchMapset)
	notifications["update"] = make(map[uint][]*OsuSearchMapset)

	processMapsets()
	notify()

	logMsg(fmt.Sprintf("Finished run for %s.", today))
}

// processMapsets gets mapsets from osusearch.com and processes them.
func processMapsets() {
	for i := 0; i < 2; i++ { // 1000 mapsets.
		mapsets, err := getMapsets(i)
		if err != nil {
			logMsg(fmt.Sprintf("Getting data from osusearch.com failed: %s", err))
		}

		mapsets = dedup(mapsets)
		log.Printf("Retrieved %d mapset(s)\n", len(mapsets))

		for _, mapset := range mapsets {
			existing := new(gosubscribe.Mapset)
			gosubscribe.DB.Where("id = ?", mapset.ID).First(existing)
			if existing.ID == 0 {
				processMapset(mapset, "new")
			} else if existing.Status != mapset.Status {
				processMapset(mapset, "status")
			} else {
				processMapset(mapset, "update")
			}
		}
	}
}

// notify sends messages to users about their subscriptions.
func notify() {
	var users []*gosubscribe.User
	gosubscribe.DB.Find(&users)

	for _, user := range users {
		userNotifications := make(map[string][]*OsuSearchMapset)
		userNotifications["new"] = notifications["new"][user.ID]
		userNotifications["status"] = notifications["status"][user.ID]
		userNotifications["update"] = notifications["update"][user.ID]

		if user.MessageOsu {
			if user.OsuUsername.Valid {
				// TODO; continue if successful, fall back to Discord otherwise.
				// Will have to deal with one-line messages, probably splitting them.
				// msg := createMessage(user, userNotifications, "osu")
				// lines := strings.split(strings.TrimSpace(msg), "\n")
				// if len(lines) == 1 { // No notifications for this user.
				// 	continue
				// }
			}
			log.Printf("User %d has MessageOsu set but not OsuUsername", user.ID)
		}

		msg := createMessage(user, userNotifications, "discord")
		fmt.Println(msg)
		if len(strings.Split(strings.TrimSpace(msg), "\n")) == 1 {
			continue // No notifications for this user.
		}

		if user.DiscordID.Valid {
			dUser, err := bot.User(strconv.Itoa(int(user.DiscordID.Int64)))
			if err != nil {
				log.Println(err)
				continue
			}

			channel, err := bot.UserChannelCreate(dUser.ID)
			if err != nil {
				logMsg(fmt.Sprintf(
					"Sending to %s failed: Couldn't open a private message channel.",
					dUser.Mention(),
				))
				continue
			}

			_, err = bot.ChannelMessageSend(channel.ID, msg)
			if err != nil {
				logMsg(fmt.Sprintf("Sending to %s failed: '%s'", dUser.Mention(), err))
			} else {
				logMsg(fmt.Sprintf(
					"Sent message to `%s#%s` (Discord).",
					dUser.Username, dUser.Discriminator,
				))
			}
		} else {
			log.Printf("Couldn't create a Discord user from user %d\n", user.ID)
		}
	}
}

func createMessage(
	user *gosubscribe.User,
	mapsets map[string][]*OsuSearchMapset,
	platform string,
) string {
	msg := fmt.Sprintf("Notifications for %s:\n", today)

	for _, mapset := range mapsets["new"] {
		fmt.Println("!!!!")
		var mapString string
		switch platform {
		case "osu":
			mapString = osuString(mapset)
		case "discord":
			mapString = discordString(mapset)
		default:
			mapString = defaultString(mapset)
		}
		msg += fmt.Sprintf("\nNew map: %s", mapString)
	}

	for _, mapset := range mapsets["status"] {
		var mapString string
		switch platform {
		case "osu":
			mapString = osuString(mapset)
		case "discord":
			mapString = discordString(mapset)
		default:
			mapString = defaultString(mapset)
		}
		msg += fmt.Sprintf(
			"\nStatus updated to %s: %s", statusMap[mapset.Status], mapString,
		)
	}

	if !user.NotifyAll {
		return msg
	}

	for _, mapset := range mapsets["update"] {
		var mapString string
		switch platform {
		case "osu":
			mapString = osuString(mapset)
		case "discord":
			mapString = discordString(mapset)
		default:
			mapString = defaultString(mapset)
		}
		msg += fmt.Sprintf("\nMap updated by mapper: %s", mapString)
	}
	return msg
}

// getSubs gets all users who are subscribed to a mapper.
func getSubs(mapper *gosubscribe.Mapper) []*gosubscribe.User {
	users := []*gosubscribe.User{}
	subs := []gosubscribe.Subscription{}
	gosubscribe.DB.Table("subscriptions").Where("mapper_id = ?", mapper.ID).Find(&subs)
	for _, sub := range subs {
		user, err := gosubscribe.GetUser(sub.UserID)
		if err != nil {
			log.Printf("No user found for %d\n", sub.UserID)
		} else {
			users = append(users, user)
		}
	}
	log.Printf("Retrieved %d subscriber(s) for %s\n", len(users), mapper.Username)
	return users
}

// processMapset updates a mapset in the DB.
func processMapset(mapset *OsuSearchMapset, key string) {
	mapper, err := gosubscribe.MapperFromDB(mapset.Mapper)
	if err != nil {
		return // No mapper means no subs.
	}

	subs := getSubs(mapper)
	for _, sub := range subs {
		log.Println("Adding to notifications")
		notifications[key][sub.ID] = append(notifications[key][sub.ID], mapset)
	}

	updated := gosubscribe.Mapset{
		ID: mapset.ID, MapperID: mapper.ID, Status: mapset.Status,
	}
	if key == "new" {
		logMsg(fmt.Sprintf("New map: %s", discordString(mapset)))
		gosubscribe.DB.Create(&updated)
	} else if key == "status" {
		logMsg(fmt.Sprintf("Ranked status updated: %s", discordString(mapset)))
		gosubscribe.DB.Save(&updated)
	} else {
		logMsg(fmt.Sprintf("Updated by mapper: %s", discordString(mapset)))
	}
}

// dedup removes maps from the same beatmapset from the list (maybe unnecessary).
func dedup(list []*OsuSearchMapset) []*OsuSearchMapset {
	uniq := []*OsuSearchMapset{}
	for _, mapset := range list {
		contains := false
		for _, existing := range uniq {
			if mapset.ID == existing.ID {
				contains = true
				break
			}
		}
		if !contains {
			uniq = append(uniq, mapset)
		}
	}
	return uniq
}

// getMapsets retrieves mapsets from osusearch.com.
func getMapsets(offset int) ([]*OsuSearchMapset, error) {
	url := fmt.Sprintf("%s?key=%s&count=500&offset=%d", searchURL, searchKey, offset)
	log.Printf("Requesting from %s\n", strings.Replace(url, searchKey, "[secure]", 1))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SearchResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Mapsets, nil
}

// logMsg logs a message to a Discord channel.
func logMsg(msg string) {
	log.Println(msg)
	bot.ChannelMessageSend(logChannel, msg)
}

// discordString converts a mapset into a stylized string for Discord.
func discordString(mapset *OsuSearchMapset) string {
	return fmt.Sprintf(
		"`%s - %s` by `%s` (<https://osu.ppy.sh/s/%d>)",
		mapset.Artist, mapset.Title, mapset.Mapper, mapset.ID,
	)
}

// osuString converts a mapset into a stylized string for IRC.
func osuString(mapset *OsuSearchMapset) string {
	return fmt.Sprintf(
		"[https://osu.ppy.sh/s/%d %s - %s] by [https://osu.ppy.sh/u/%s %s]",
		mapset.ID, mapset.Artist, mapset.Title, mapset.Mapper, mapset.Mapper,
	)
}

// defaultString converts a mapset into a string with no styling.
func defaultString(mapset *OsuSearchMapset) string {
	return fmt.Sprintf("%s - %s by %s", mapset.Artist, mapset.Title, mapset.Mapper)
}
