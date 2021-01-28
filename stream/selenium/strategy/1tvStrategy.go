package strategy

import (
	"fmt"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	"log"
	"time"
)

func Fetch1tvManifest(wd selenium.WebDriver) string {
	fmt.Printf("-----go to %s site-----\n", channel.ChUrls[channel.FIRST])
	if err := wd.Get(channel.ChUrls[channel.FIRST]); err != nil {
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
