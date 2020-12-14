package stream

import (
	"errors"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	seleLog "github.com/tebeka/selenium/log"
	"log"
	"regexp"
	"time"
)

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

	return selenium.NewRemote(caps, "http://selenium:4444/wd/hub")
}

func GetManifest() string {
	wd, err := createWebDriver()
	if err != nil {
		log.Fatalf("cant create wevDriver, error: %v", err)
	}

	defer func() {
		if err := wd.Quit(); err != nil {
			log.Fatalf("error while quit webDriver, error: %v", err)
		}
	}()

	link, err := fetchTntManifest(wd)
	if err != nil {
		log.Fatalln(err)
	}
	return link
}
func fetchTntManifest(wd selenium.WebDriver) (string, error) {
	if err := wd.Get("https://tnt-online.ru/live"); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}

	time.Sleep(time.Second * 2)
	elem, err := wd.FindElement(selenium.ByClassName, "live__frame")
	if err != nil {
		log.Fatalf("not found iframe tag, %v", err)
	}

	src, err := elem.GetAttribute("src")
	if err != nil {
		log.Fatalf("not found src attribute, %v", err)
	}

	if err := wd.Get(src); err != nil {
		log.Fatalf("error while fetch url: %v", err)
	}
	time.Sleep(time.Second * 5)

	return findSourceAtLogs(wd)
}
func findSourceAtLogs(wd selenium.WebDriver) (string, error) {
	message, err := wd.Log(seleLog.Performance)
	if err != nil {
		log.Fatalf("error get log, %s", err)
	}

	re := regexp.MustCompile(`https:\/\/live(.)+(\.m3u8)`)
	for _, v := range message {
		if found := re.Find([]byte(v.Message)); found != nil {
			return string(found), nil
		}
	}

	return "", errors.New("live file is not found in logs")
}
