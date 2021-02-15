package strategy

import (
	"fmt"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	"log"
	"time"
)

func FetchTntManifest(wd selenium.WebDriver, pattern string) string {
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
	//makeScreenshot(wd, channel.TNT)

	fmt.Println("-----searching manifest-----")
	if pattern == "" {
		pattern = `https:\/\/live(.)+(\.m3u8)`
	}

	link, err := findSourceAtLogs(wd, pattern)
	if err != nil {
		fmt.Println(err)
		link = FetchTntManifest(wd, `https:\/\/matchtv(.)+(\.m3u8)`)
	}
	fmt.Println("-----got link-----")

	return link
}

type Pattern struct {
	pattern string
	attempt int
}

func getPattern() *Pattern {
	var list []*Pattern
	pattern1 := &Pattern{
		pattern: `https:\/\/live(.)+(\.m3u8)`,
		attempt: 0,
	}
	pattern2 := &Pattern{
		pattern: `https:\/\/matchtv(.)+(\.m3u8)`,
		attempt: 0,
	}
	list = append(list, pattern1, pattern2)

	needle := &Pattern{attempt: 0}
	for _, v := range list {
		if needle.attempt <= v.attempt {
			continue
		}
		needle = v
	}
	return needle
}
