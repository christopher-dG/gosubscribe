package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/christopher-dG/gosubscribe"
)

var (
	statusMap = map[int]string{
		-1: "WIP",
		0:  "pending",
		1:  "ranked",
		2:  "approved",
		3:  "qualified",
		4:  "loved",
	}
	logChannel    = os.Getenv("DISCORD_LOG_CHANNEL")
	searchKey     = os.Getenv("OSUSEARCH_API_KEY")
	searchURL     = "http://osusearch.com/api/search"
	notifications = make(map[string]map[*gosubscribe.User]*OsuSearchMapset)
	bot           *discordgo.Session
)

// SearchResult is the JSON structure returned by osusearch.com.
type SearchResult struct {
	Maps []*OsuSearchMapset `json:"beatmaps"`
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

	today := time.Now().Format("2006-01-02")
	logMsg(fmt.Sprintf("Starting run for %s...", today))

	notifications["new"] = make(map[*gosubscribe.User]*OsuSearchMapset)
	notifications["statusUpdate"] = make(map[*gosubscribe.User]*OsuSearchMapset)
	notifications["update"] = make(map[*gosubscribe.User]*OsuSearchMapset)

	for i := 0; i < 2; i++ { // 1000 maps.
		mapsets, err := getMapsets(i)
		if err != nil {
			logMsg(fmt.Sprintf("Getting data from osusearch.com failed: %s", err))
		}

		mapsets = dedup(mapsets)

		for _, mapset := range mapsets {
			existing := new(gosubscribe.Mapset)
			gosubscribe.DB.Where("id = ?", mapset.ID).First(existing)
			if existing.ID == 0 {
				// processNew(mapset)
			} else if existing.Status == mapset.Status {
				// processStatusUpdate(mapset)
			} else {
				// processUpdate(mapset)
			}
		}

	}
	logMsg(fmt.Sprintf("Finished run for %s.", today))
}

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
	log.Printf("Retrieved %d subscribers for %s\n", len(users), mapper.Username)
	return users
}

func processNew(mapset *OsuSearchMapset) {
	mapper, err := gosubscribe.MapperFromDB(mapset.Mapper)
	if err != nil {
		return // No mapper means no subs.
	}
	subs := getSubs(mapper)
	for _, sub := range subs {
		notifications["new"][sub] = mapset
	}
}

func processStatusUpdate(mapset *OsuSearchMapset) {
	mapper, err := gosubscribe.MapperFromDB(mapset.Mapper)
	if err != nil {
		return // No mapper means no subs.
	}
	subs := getSubs(mapper)
	for _, sub := range subs {
		notifications["statusUpdate"][sub] = mapset
	}
}

func processUpdate(mapset *OsuSearchMapset) {
	mapper, err := gosubscribe.MapperFromDB(mapset.Mapper)
	if err != nil {
		return // No mapper means no subs.
	}
	subs := getSubs(mapper)
	for _, sub := range subs {
		notifications["update"][sub] = mapset
	}
}

// dedup removes maps from the same beatmapset from the list (maybe unnecessary).
func dedup(list []*OsuSearchMapset) []*OsuSearchMapset {
	uniq := []*OsuSearchMapset{}
	for _, mapset := range list {
		contains := false
		fmt.Println(mapset.ID)
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

	maps := []*OsuSearchMapset{}
	for i, _ := range result.Maps {
		maps = append(maps, result.Maps[i])
	}

	return maps, nil
}

func logMsg(msg string) {
	log.Println(msg)
	// bot.ChannelMessageSend(logChannel, msg)
}
