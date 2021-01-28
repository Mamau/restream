package selenium

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/mamau/restream/stream/selenium/strategy"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	seleLog "github.com/tebeka/selenium/log"
	"log"
	"sync"
)

var once sync.Once

type Selenium struct {
	wd selenium.WebDriver
}

var instance *Selenium

func NewSelenium() *Selenium {
	once.Do(func() {
		wd, err := createWebDriver()
		if err != nil {
			log.Fatalf("cant create wevDriver, error: %v", err)
		}
		instance = &Selenium{
			wd: wd,
		}
	})

	return instance
}

func createWebDriver() (selenium.WebDriver, error) {
	fmt.Println("-----create selenium chrome driver-----")
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			"--headless",
			"--no-sandbox",
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
		},
	}

	caps.AddChrome(chromeCaps)
	lg := make(map[seleLog.Type]seleLog.Level)
	lg[seleLog.Performance] = seleLog.Info
	caps.AddLogging(lg)

	return selenium.NewRemote(caps, "http://selenium:4444/wd/hub")
	//return selenium.NewRemote(caps, "http://0.0.0.0:4444/wd/hub")
}

func GetManifest(ch channel.Channel) (string, error) {
	s := NewSelenium()
	wd := s.wd

	defer func() {
		fmt.Println("-----selenium quit-----")
		if err := wd.Quit(); err != nil {
			log.Fatalf("error while quit webDriver, error: %v", err)
		}
	}()

	switch ch {
	case channel.TNT:
		return strategy.FetchTntManifest(wd), nil
	case channel.FIRST:
		return strategy.Fetch1tvManifest(wd), nil
	case channel.MATCH:
		return strategy.FetchMatchManifest(wd), nil
	default:
		return "", errors.New(fmt.Sprintf("Unknown type channel: %s", ch))
	}
}
