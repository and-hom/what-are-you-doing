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
	GetWorkingHourForToday() int
}

func snapshotToString(report map[string]int, week int) string {
	var buffer bytes.Buffer

	local, _ := time.LoadLocation("Local")
	monday := firstDayOfISOWeek(time.Now().Year(), week, local)
	friday := monday.Add((5 * 24 - 1) * time.Hour)
	buffer.WriteString(fmt.Sprintf("====%s - %s\n", monday.Format("2006-01-02"), friday.Format("2006-01-02")))
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
func (v StdoutJobLogger) GetWorkingHourForToday() int {
	return 0
}

const CSV_SEPARATOR = "\t"
const MIN_IDLE_TIME_FOR_NEW_DAY_DETECTION_HOURS int = 8
var STARTED_AT int64 = time.Now().Unix()


func CreateFileJobLogger(baseDir string) FileJobLogger {
	jobLogger := FileJobLogger{
		Basedir: baseDir,
		currentWeekDump:[]int64{},
	}
	jobLogger.scan(jobLogger.thisWeek(), func(ts int64, _ string){
		jobLogger.currentWeekDump = append(jobLogger.currentWeekDump, ts)
	})
	return jobLogger
}

type FileJobLogger struct {
	currentWeek     int;
	Basedir         string;
	currentWeekDump []int64;
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
	timestamp := time.Now().Unix()
	_, e = file.WriteString(fmt.Sprintf("%d%s%s\n", timestamp, CSV_SEPARATOR, project))
	if e != nil {
		log.Panic(e)
	}
	v.currentWeekDump = append(v.currentWeekDump, timestamp)
}

func (v FileJobLogger) scan(week int, callback func(ts int64, project string)) {
	path := v.logPath(week)
	if _, err := os.Stat(path); err == nil {
		snapshotFile, e := os.Open(path)
		if e != nil {
			log.Panicf("Can not open file: %s", e)
		}
		defer snapshotFile.Close()

		scanner := bufio.NewScanner(snapshotFile)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), CSV_SEPARATOR)
			ts, e := strconv.ParseInt(parts[0], 10, 64)
			if e != nil {
				continue
			}
			callback(ts, parts[1])
		}
	}
}

func (v FileJobLogger) weekSnapshot(week int) (map[string]int, int) {
	result := map[string]int{}
	sum := 0

	v.scan(week, func(_ int64, project string){
		result[project] += 1
		sum += 1
	})

	if sum > 0 {
		for k, _ := range result {
			result[k] = result[k] * 100 / sum;
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

func (v FileJobLogger) GetWorkingHourForToday() int {
	return deltaNowHours(v.firstTimestampForToday())
}

func (v FileJobLogger) firstTimestampForToday() int64 {
	length := len(v.currentWeekDump)
	if length == 0 {
		return STARTED_AT
	}
	if  deltaNowHours(v.currentWeekDump[length - 1])> MIN_IDLE_TIME_FOR_NEW_DAY_DETECTION_HOURS {
		return STARTED_AT
	}
	if length == 1 {
		return v.currentWeekDump[0]
	}
	for i := length - 2; i >= 0; i-- {
		if deltaHours(v.currentWeekDump[i + 1], v.currentWeekDump[i]) > MIN_IDLE_TIME_FOR_NEW_DAY_DETECTION_HOURS {
			return v.currentWeekDump[i + 1]
		}
	}
	return v.currentWeekDump[0]
}

func deltaNowHours(ts int64) int {
	return deltaHours(time.Now().Unix() ,ts)
}
func deltaHours(ts2 int64, ts1 int64) int {
	return int((ts2 - ts1) / int64(3600))
}

