package process

import (
	"fmt"
	"log"
	"os"
	"time"
)

var lastmodFormat string = "%s.lastmod"
var dateFormat = "2006-01-02 15:04:05.000"
var defaultDate = "1900-01-01 00:00:00.000"
var DefaultModTime time.Time

func init() {
	DefaultModTime, _ = time.Parse(dateFormat, defaultDate)
	DefaultModTime = DefaultModTime.UTC()
}

func ReadLastMod(cfgfilename string) time.Time {
	b, err := os.ReadFile(fmt.Sprintf(lastmodFormat, cfgfilename))
	if err != nil {
		return DefaultModTime
	}

	s := string(b)
	dt, err := time.Parse(dateFormat, s)
	if err != nil {
		return DefaultModTime
	}
	return dt
}

func WriteLastMod(cfgfilename string) {
	dt := time.Now().UTC()
	f := dt.Format(dateFormat)
	err := os.WriteFile(fmt.Sprintf(lastmodFormat, cfgfilename), []byte(f), os.ModePerm)
	if err != nil {
		log.Println("problem writing last mod file", err.Error())
	}
}
