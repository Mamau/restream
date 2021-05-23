package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

const (
	TB = 1000000000000
	GB = 1000000000
	MB = 1000000
	KB = 1000
)

func ByteHuman(length int, decimals int) (out string) {
	var unit string
	var i int
	var remainder int

	if length > TB {
		unit = "TB"
		i = length / TB
		remainder = length - (i * TB)
	} else if length > GB {
		unit = "GB"
		i = length / GB
		remainder = length - (i * GB)
	} else if length > MB {
		unit = "MB"
		i = length / MB
		remainder = length - (i * MB)
	} else if length > KB {
		unit = "KB"
		i = length / KB
		remainder = length - (i * KB)
	} else {
		return strconv.Itoa(length) + " B"
	}

	if decimals == 0 {
		return strconv.Itoa(i) + " " + unit
	}

	width := 0
	if remainder > GB {
		width = 12
	} else if remainder > MB {
		width = 9
	} else if remainder > KB {
		width = 6
	} else {
		width = 3
	}

	remainderString := strconv.Itoa(remainder)
	for iter := len(remainderString); iter < width; iter++ {
		remainderString = "0" + remainderString
	}
	if decimals > len(remainderString) {
		decimals = len(remainderString)
	}

	return fmt.Sprintf("%d.%s %s", i, remainderString[:decimals], unit)
}

func CyrillicToLatin(name string) string {
	var translated []string
	dict := map[string]string{
		"а": "a",
		"б": "b",
		"в": "v",
		"г": "g",
		"д": "d",
		"е": "e",
		"ё": "e",
		"ж": "zh",
		"з": "z",
		"и": "i",
		"й": "i",
		"к": "k",
		"л": "l",
		"м": "m",
		"н": "n",
		"о": "o",
		"п": "p",
		"р": "r",
		"с": "s",
		"т": "t",
		"у": "u",
		"ф": "f",
		"х": "h",
		"ц": "ts",
		"ч": "ch",
		"ш": "sh",
		"щ": "sh",
		"ъ": "",
		"ы": "i",
		"ь": "",
		"э": "e",
		"ю": "u",
		"я": "ia",
		" ": "_",
	}

	for _, v := range strings.Split(name, "") {
		if val, ok := dict[strings.ToLower(v)]; ok {
			v = val
		}

		translated = append(translated, v)
	}

	return strings.ToLower(strings.Join(translated, ""))
}

func CurrentDir() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cant get current directory")
	}
	return pwd
}
