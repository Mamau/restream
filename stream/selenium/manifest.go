package selenium

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/mamau/restream/stream/selenium/strategy"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	seleLog "github.com/tebeka/selenium/log"
)

type Selenium struct {
	wd selenium.WebDriver
}

func NewSelenium() *Selenium {
	wd, err := createWebDriver()
	if err != nil {
		log.Fatalf("cant create wevDriver, error: %v", err)
	}
	return &Selenium{wd}
}

func createWebDriver() (selenium.WebDriver, error) {
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

	return selenium.NewRemote(caps, os.Getenv("SELENIUM"))
}

func GetManifest(ch channel.Channel) (string, error) {
	if isTest, _ := strconv.ParseBool(os.Getenv("IS_TEST")); isTest {
		return "https://ya.ru", nil
	}
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
