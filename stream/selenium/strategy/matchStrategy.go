package strategy

import (
	"fmt"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	"log"
	"time"
)

func FetchMatchManifest(wd selenium.WebDriver) string {
	fmt.Printf("-----go to %s site-----\n", channel.ChUrls[channel.MATCH])
	if err := wd.Get(channel.ChUrls[channel.MATCH]); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}
	fmt.Println("-----wait 3 sec-----")
	time.Sleep(time.Second * 3)

	//makeScreenshot(wd, channel.MATCH)

	elem, err := wd.FindElement(selenium.ByClassName, "video-player")

	if err != nil {
		log.Fatalf("not found id videoShare")
	}

	if err := elem.Click(); err != nil {
		fmt.Println("-----Cant click, try again-----")
		FetchMatchManifest(wd)
	}

	fmt.Println("-----wait 3 again sec-----")
	time.Sleep(time.Second * 3)

	//makeScreenshot(wd, channel.MATCH)
	fmt.Println("-----search manifest-----")
	link, err := findSourceAtLogs(wd, `https:\/\/live(.)+(\.m3u8)`)
	if err != nil {
		fmt.Println(err)
		link = FetchMatchManifest(wd)
	}
	fmt.Println("-----got link-----", link)

	return link
}
