package main

import (
	log "github.com/Sirupsen/logrus"
	"fmt"
	"os"
	"time"
	"bufio"
	"strings"
	"strconv"
	"bytes"
)

type JobLogger interface {
	AddForNow(project string)
	// percent of time by project
	PrviousWeekSnapshot() (map[string]int, int)
	ThisWeekSnapshot() (map[string]int, int)
}


func snapshotToString(report map[string]int, week int) string {
	var buffer bytes.Buffer

	local, _ := time.LoadLocation("Local")
	monday := firstDayOfISOWeek(time.Now().Year(), week, local)
	friday := monday.Add(5 * 24 * time.Hour)
	buffer.WriteString(fmt.Sprintf("%s - %s\n", monday.Format("2006-01-02"), friday.Format("2006-01-02")))
	for k, v := range report {
		buffer.WriteString(fmt.Sprintf("**%s** - %d%%\n", k, v))
	}

	return buffer.String()
}

type StdoutJobLogger struct {
}

func (v StdoutJobLogger) AddForNow(project string) {
	fmt.Printf("%s\n", project)
}
func (v StdoutJobLogger) PrviousWeekSnapshot() (map[string]int, int) {
	return nil, 0
}
func (v StdoutJobLogger) ThisWeekSnapshot() (map[string]int, int) {
	return nil, 0
}

const CSV_SEPARATOR = "\t"

type FileJobLogger struct {
	currentWeek int;
	Basedir     string;
}

func (v FileJobLogger) thisWeek() int {
	_, w := time.Now().ISOWeek()
	return int(w)
}

func (v FileJobLogger) logPath(week int) string {
	return fmt.Sprintf("%s/%d-%d.txt", v.Basedir, time.Now().Year(), week)
}

func (v FileJobLogger) summaryPath() string {
	return fmt.Sprintf("%s/%d-summary.txt", v.Basedir, v.thisWeek())
}

func (v FileJobLogger) AddForNow(project string) {
	os.MkdirAll(v.Basedir, 0777)
	path := v.logPath(v.thisWeek())
	file, e := os.OpenFile(path, os.O_RDWR | os.O_APPEND | os.O_CREATE, 0660);
	if e != nil {
		log.Panic(e)
	}
	defer file.Close()
	_, e = file.WriteString(fmt.Sprintf("%d%s%s\n", time.Now().Unix(), CSV_SEPARATOR, project))
	if e != nil {
		log.Panic(e)
	}
}

func (v FileJobLogger) weekSnapshot(week int) (map[string]int, int) {
	path := v.logPath(week)
	result := map[string]int{}

	if _, err := os.Stat(path); err == nil {
		snapshotFile, e := os.Open(path)
		if e != nil {
			log.Panicf("Can not open file: %s", e)
		}
		defer snapshotFile.Close()

		scanner := bufio.NewScanner(snapshotFile)
		scanner.Split(bufio.ScanLines)

		sum := 0
		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), CSV_SEPARATOR)
			_, e := strconv.ParseInt(parts[0], 10, 64)
			if e != nil {
				continue
			}
			project := parts[1]
			result[project] += 1
			sum += 1
		}

		if sum > 0 {
			for k, _ := range result {
				result[k] = result[k] * 100 / sum;
			}
		}
	}
	return result, week
}

func (v FileJobLogger) PrviousWeekSnapshot() (map[string]int, int) {
	return v.weekSnapshot(v.thisWeek() - 1)
}

func (v FileJobLogger) ThisWeekSnapshot() (map[string]int, int) {
	return v.weekSnapshot(v.thisWeek())
}

