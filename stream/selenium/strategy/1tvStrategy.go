package strategy

import (
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"time"
)

func Fetch1tvManifest(wd selenium.WebDriver) string {
	url := "https://www.1tv.ru/live"
	fmt.Printf("-----go to %s site-----\n", url)
	if err := wd.Get(url); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}
	fmt.Println("-----wait 3 sec-----")
	time.Sleep(time.Second * 3)

	fmt.Println("-----searching manifest-----")
	link, err := findSourceAtLogs(wd, `https:\/\/edge(.)+(\.mpd\?[a-z]{1}\=[0-9]+)`)
	if err != nil {
		fmt.Println(err)
		link = Fetch1tvManifest(wd)
	}
	fmt.Println("-----got link-----")

	return link
}
