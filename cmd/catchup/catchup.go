package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/christopher-dG/gosubscribe"
	irc "github.com/thoj/go-ircevent"
)

var (
	bot          *irc.Connection
	namesChannel = make(chan string)
)

func main() {
	gosubscribe.Connect(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASS"),
	)

	bot = irc.IRC(os.Getenv("IRC_USER"), os.Getenv("IRC_USER"))
	bot.Password = os.Getenv("IRC_PASS")
	err := bot.Connect(fmt.Sprintf("%s:%s", os.Getenv("IRC_SERVER"), os.Getenv("IRC_PORT")))
	if err != nil {
		log.Fatal(err)
	}
	bot.AddCallback("353", func(e *irc.Event) { namesChannel <- e.Message() })
	for {
		time.Sleep(3 * time.Second)
		fmt.Printf("%d\n", len(getOnline()))
	}
}

func getOnline() []string {
	names := []string{}
	bot.SendRaw("NAMES #osu")
	time.Sleep(time.Second)
	for len(namesChannel) > 0 {
		line := <-namesChannel
		fmt.Println(line)
		for _, name := range strings.Split(line, " ") {
			names = append(names, name)
		}
	}
	return names
}
