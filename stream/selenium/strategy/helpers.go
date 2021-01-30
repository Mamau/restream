package strategy

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/tebeka/selenium"
	seleLog "github.com/tebeka/selenium/log"
	"log"
	"os"
	"regexp"
	"time"
)

func findSourceAtLogs(wd selenium.WebDriver, pattern string) (string, error) {
	message, err := wd.Log(seleLog.Performance)
	if err != nil {
		log.Fatalf("error get log, %s", err)
	}

	re := regexp.MustCompile(pattern)
	for _, v := range message {
		if found := re.Find([]byte(v.Message)); found != nil {
			return string(found), nil
		}
	}

	return "", errors.New("live file is not found in logs")
}

func makeScreenshot(wd selenium.WebDriver, channel channel.Channel) {
	folder := fmt.Sprintf("%v/%v/%v", helpers.CurrentDir(), "storage/logs", channel)
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		log.Fatalf("cant create folder %s\n", folder)
	}

	imgBytes, err := wd.Screenshot()
	if err != nil {
		fmt.Println("cant get screenshot")
	}

	screenShotName := fmt.Sprintf("%s/screenshot_%s.png", folder, time.Now().Format("15_04_05"))
	f, err := os.OpenFile(screenShotName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("error while open screenshot file for write, err: %s\n", err)
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
