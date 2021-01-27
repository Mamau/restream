package strategy

import (
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	seleLog "github.com/tebeka/selenium/log"
	"log"
	"os"
	"regexp"
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

func makeScreenshot(wd selenium.WebDriver) {
	imgBytes, err := wd.Screenshot()
	if err != nil {
		fmt.Println("cant get screenshot")
	}

	if e := os.Remove("screenshot.png"); e != nil {
		fmt.Println("no screenshot file or cant delete it")
	}

	f, err := os.OpenFile("screenshot.png", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
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
