package helpers

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

func JsonRequestToMap(r *http.Request, s interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(s)
	if err != nil {
		return err
	}
	return nil
}

func Pwd() string {
	data, _ := exec.Command("pwd").CombinedOutput()
	pwd := strings.TrimRight(string(data), "\n")
	return pwd
}
