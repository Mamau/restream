package strategy

import (
	"fmt"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	"log"
	"time"
)

func FetchTntManifest(wd selenium.WebDriver) string {
	fmt.Printf("-----go to %s site-----\n", channel.ChUrls[channel.TNT])
	if err := wd.Get(channel.ChUrls[channel.TNT]); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}

	fmt.Println("-----wait 2 sec-----")
	time.Sleep(time.Second * 2)

	fmt.Println("-----searching live__frame class-----")
	elem, err := wd.FindElement(selenium.ByClassName, "live__frame")
	if err != nil {
		log.Fatalf("not found iframe tag, %v", err)
	}

	fmt.Println("-----get src tag-----")
	src, err := elem.GetAttribute("src")
	if err != nil {
		log.Fatalf("not found src attribute, %v", err)
	}

	fmt.Printf("-----go to %s-----\n", src)
	if err := wd.Get(src); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}
	fmt.Println("-----wait 5 sec-----")
	time.Sleep(time.Second * 5)

	fmt.Println("-----searching manifest-----")
	link, err := findSourceAtLogs(wd, `https:\/\/live(.)+(\.m3u8)`)
	if err != nil {
		fmt.Println(err)
		link = FetchTntManifest(wd)
	}
	fmt.Println("-----got link-----")

	return link
}
