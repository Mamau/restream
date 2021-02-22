package strategy

import (
	"fmt"
	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	"time"
)

func Fetch1tvManifest(wd selenium.WebDriver) string {
	fmt.Printf("-----go to %s site-----\n", channel.ChUrls[channel.FIRST])
	if err := wd.Get(channel.ChUrls[channel.FIRST]); err != nil {
		storage.GetLogger().Fatal(err)
	}
	fmt.Println("-----wait 3 sec-----")
	time.Sleep(time.Second * 3)

	fmt.Println("-----searching manifest-----")
	link, err := findSourceAtLogs(wd, channel.ChanManifestPatterns[channel.FIRST])

	if err != nil {
		storage.GetLogger().Fatal(err)
	}
	fmt.Println("-----got link-----")

	return link
}
