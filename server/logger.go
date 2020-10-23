package server

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

var SugarLogger *zap.SugaredLogger

func InitLogger() {
	data, _ := exec.Command("pwd").CombinedOutput()
	pwd := strings.Split(string(data), "\n")
	res := strings.Join(append(pwd, "/logs/log.txt"), "")

	cnf := fmt.Sprintf(
		`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["%v"],
	  "errorOutputPaths": ["%v"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`, res, res)
	rawJSON := []byte(cnf)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	SugarLogger = logger.Sugar()
}
