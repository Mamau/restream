package strategy

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	seleLog "github.com/tebeka/selenium/log"
	"os"
	"regexp"
	"time"
)

func findSourceAtLogs(wd selenium.WebDriver, patterns []*channel.Pattern) (string, error) {
	message, err := wd.Log(seleLog.Performance)
	if err != nil {
		storage.GetLogger().Fatal(err)
	}

	for _, v := range message {
		for _, p := range patterns {
			re := regexp.MustCompile(p.Scheme)
			if found := re.Find([]byte(v.Message)); found != nil {
				return string(found), nil
			}
		}
	}
	storage.GetLogger().Info(fmt.Sprintf("no find manifest in heap: %s", message))
	return "", errors.New("live file is not found in logs")
}

func makeScreenshot(wd selenium.WebDriver, channel channel.Channel) {
	folder := fmt.Sprintf("%v/%v/%v", helpers.CurrentDir(), "storage/logs", channel)
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		storage.GetLogger().Fatal(err)
	}

	imgBytes, err := wd.Screenshot()
	if err != nil {
		fmt.Println("cant get screenshot")
	}

	screenShotName := fmt.Sprintf("%s/screenshot_%s.png", folder, time.Now().Format("15_04_05"))
	f, err := os.OpenFile(screenShotName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		storage.GetLogger().Fatal(err)
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
