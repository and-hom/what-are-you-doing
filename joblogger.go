package main

import (
	log "github.com/Sirupsen/logrus"
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"time"
)

type JobLogger interface {
	AddForNow(project string)
	// percent of time by project
	PrviousWeekSnapshot() map[string]int
	ThisWeekSnapshot() map[string]int
}

type StdoutJobLogger struct {
}

func (v StdoutJobLogger) AddForNow(project string) {
	fmt.Printf("%s\n", project)
}
func (v StdoutJobLogger) PrviousWeekSnapshot() map[string]int {
	return nil
}
func (v StdoutJobLogger) ThisWeekSnapshot() map[string]int {
	return nil
}

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
		log.Fatal(e)
	}
	defer file.Close()
	_, e = file.WriteString(fmt.Sprintf("%d\t%s\n", time.Now().Unix(), project))
	if e != nil {
		log.Fatal(e)
	}
}
func (v FileJobLogger) weekSnapshot(week int) map[string]int {
	path := v.logPath(week)
	content, e := ioutil.ReadFile(path)
	if e != nil {
		log.Fatal(e)
	}
	lines := strings.Split(string(content), "\n")
	result := map[string]int{}
	for _, line := range lines {
		result[line] += 1
	}
	return result
}
func (v FileJobLogger) PrviousWeekSnapshot() map[string]int {
	return v.weekSnapshot(v.thisWeek() - 1)
}
func (v FileJobLogger) ThisWeekSnapshot() map[string]int {
	return v.weekSnapshot(v.thisWeek())
}

