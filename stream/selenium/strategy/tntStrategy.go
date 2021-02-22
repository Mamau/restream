package strategy

import (
	"fmt"
	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	"time"
)

func FetchTntManifest(wd selenium.WebDriver) string {
	fmt.Printf("-----go to %s site-----\n", channel.ChUrls[channel.TNT])
	if err := wd.Get(channel.ChUrls[channel.TNT]); err != nil {
		storage.GetLogger().Fatal(err)
	}

	fmt.Println("-----wait 2 sec-----")
	time.Sleep(time.Second * 2)

	fmt.Println("-----searching live__frame class-----")
	elem, err := wd.FindElement(selenium.ByClassName, "live__frame")
	if err != nil {
		storage.GetLogger().Fatal(err)
	}

	fmt.Println("-----get src tag-----")
	src, err := elem.GetAttribute("src")
	if err != nil {
		storage.GetLogger().Fatal(err)
	}

	fmt.Printf("-----go to %s-----\n", src)
	if err := wd.Get(src); err != nil {
		storage.GetLogger().Fatal(err)
	}
	fmt.Println("-----wait 5 sec-----")
	time.Sleep(time.Second * 5)
	//makeScreenshot(wd, channel.TNT)

	fmt.Println("-----searching manifest-----")

	link, err := findSourceAtLogs(wd, channel.ChanManifestPatterns[channel.TNT])
	if err != nil {
		storage.GetLogger().Fatal(err)
	}
	fmt.Println("-----got link-----")

	return link
}
