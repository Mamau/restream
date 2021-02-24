package strategy

import (
	"fmt"
	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	"time"
)

func FetchMatchManifest(wd selenium.WebDriver) string {
	fmt.Printf("-----go to %s site-----\n", channel.ChUrls[channel.MATCH])
	if err := wd.Get(channel.ChUrls[channel.MATCH]); err != nil {
		storage.GetLogger().Fatal(err)
	}
	fmt.Println("-----wait 3 sec-----")
	time.Sleep(time.Second * 3)

	//makeScreenshot(wd, channel.MATCH)

	elem, err := wd.FindElement(selenium.ByClassName, "video-player")

	if err != nil {
		storage.GetLogger().Fatal(err)
	}

	if err := elem.Click(); err != nil {
		fmt.Println("-----Cant click, try again-----")
		FetchMatchManifest(wd)
	}

	fmt.Println("-----wait 3 again sec-----")
	time.Sleep(time.Second * 3)

	//makeScreenshot(wd, channel.MATCH)
	fmt.Println("-----search manifest-----")
	link, err := findSourceAtLogs(wd, channel.ChanManifestPatterns[channel.MATCH])
	if err != nil {
		storage.GetLogger().Fatal(err)
	}
	fmt.Println("-----got link-----", link)

	return link
}
