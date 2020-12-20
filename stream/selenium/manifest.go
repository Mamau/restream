package selenium

import (
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	seleLog "github.com/tebeka/selenium/log"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
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
	s := NewSelenium()
	wd := s.wd

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

	fmt.Println("go to tnt site and wait 2 sec")
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
	fmt.Println("wait 5 sec")
	time.Sleep(time.Second * 5)
	//makeScreenshot(wd)

	link, err := findSourceAtLogs(wd)
	if err != nil {
		fmt.Println(err)
		link, err = fetchTntManifest(wd)
	}
	return link, nil
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
func makeScreenshot(wd selenium.WebDriver) {
	imgBytes, err := wd.Screenshot()
	if err != nil {
		fmt.Println("cant get screenshot")
	}
	e := os.Remove("screenshot.png")
	if e != nil {
		fmt.Println("no screenshot file or cant delete it")
	}

	f, err := os.OpenFile("screenshot.png", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("cant close screenshot file")
		}
	}()
	if _, err := f.Write(imgBytes); err != nil {
		fmt.Println("cant write screenshot")
	}
}
