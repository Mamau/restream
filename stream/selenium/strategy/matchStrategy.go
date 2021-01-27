package strategy

import (
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"regexp"
	"time"
)

func FetchMatchManifest(wd selenium.WebDriver) string {
	url := "https://matchtv.ru/on-air"
	fmt.Printf("-----go to %s site-----\n", url)
	if err := wd.Get(url); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}
	fmt.Println("-----wait 3 sec-----")
	time.Sleep(time.Second * 3)

	fmt.Println("-----searching sharing tag info-----")
	elem, err := wd.FindElement(selenium.ByID, "videoShare")
	if err != nil {
		log.Fatalf("not found id videoShare")
	}

	fmt.Println("-----get value-----")
	text, err := elem.GetAttribute("value")
	if err != nil {
		log.Fatalf("not found text")
	}

	fmt.Println("-----search link-----")
	re := regexp.MustCompile(`https:\/\/([a-z]+\.[a-z]+\/[a-z]+\/[a-z]+\/[a-z0-9]+)`)
	src := re.Find([]byte(text))
	if src == nil {
		log.Fatalf("not found url in text")
	}

	fmt.Printf("-----go to %s-----\n", string(src))
	if err := wd.Get(string(src)); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}
	fmt.Println("-----wait 5 sec-----")
	time.Sleep(time.Second * 5)

	fmt.Println("-----search manifest-----")
	link, err := findSourceAtLogs(wd, `https:\/\/live(.)+(\.m3u8)`)
	if err != nil {
		fmt.Println(err)
		link = FetchMatchManifest(wd)
	}
	fmt.Println("-----got link-----")

	return link
}
