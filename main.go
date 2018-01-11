package main

import (
	log "github.com/Sirupsen/logrus"

	"fmt"
	"os"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/qml"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
)


type QmlBride struct {
	core.QObject
	_            func(project string) `slot:"okPressed"`
	_            func(normalized bool) `slot:"copyThisWeekPressed"`
	_            func(normalized bool) `slot:"copyPrevWeekPressed"`
	_            func() `slot:"windowClosed"`
	_            func() uint64 `slot:"showPeriod"`
	_            func() bool `slot:"isYellowMode"`
	_            func() bool `slot:"isRedMode"`
	_            func() `constructor:"init"`

	jobLogger    JobLogger
	clipboard    *gui.QClipboard
	showPeriodMs uint64
	workingTimeYellowLimit int
	workingTimeRedLimit int
}

func (bridge *QmlBride) copy(s string) {
	bridge.clipboard.SetText(s, gui.QClipboard__Clipboard)
}

func (bridge *QmlBride) init() {
	bridge.ConnectOkPressed(func(project string) {
		bridge.jobLogger.AddForNow(project)
		fmt.Println(" go: OK!" + project)
	})
	bridge.ConnectWindowClosed(func() {
		report, _ := bridge.jobLogger.ThisWeekSnapshot(true)
		for key, value := range report {
			fmt.Println("Key:", key, "Value:", value)
		}
	})
	bridge.ConnectCopyThisWeekPressed(func(normalized bool) {
		bridge.copy(snapshotToString(bridge.jobLogger.ThisWeekSnapshot(normalized)))
	})
	bridge.ConnectCopyPrevWeekPressed(func(normalized bool) {
		bridge.copy(snapshotToString(bridge.jobLogger.PrviousWeekSnapshot(normalized)))
	})
	bridge.ConnectShowPeriod(func() uint64 {
		return bridge.showPeriodMs
	})
	bridge.ConnectIsRedMode(func() bool {
		workingHour := bridge.jobLogger.GetWorkingHourForToday()
		return workingHour>0 && bridge.workingTimeRedLimit>0 && workingHour >= bridge.workingTimeRedLimit
	})
	bridge.ConnectIsYellowMode(func() bool {
		workingHour := bridge.jobLogger.GetWorkingHourForToday()
		return workingHour>0 && bridge.workingTimeYellowLimit>0 && workingHour >= bridge.workingTimeYellowLimit
	})

	var obj = core.NewQObject(nil)
	obj.SetObjectName("objectName")
}

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp:true, TimestampFormat:"2006-01-02 15:04:05 -0700"})
	log.SetOutput(ioutil.Discard)
	log.Info("Starting...")

	cliApp := cli.NewApp()
	cliApp.Name = "What are you doing now?"
	cliApp.Usage = "Simple task logger"

	configuration := load("")
	log.Infof("Loaded configuration %+v", configuration)

	jobLogger := CreateFileJobLogger(configuration.LogPath, 60 / configuration.AskPeriodMin)

	cliApp.Commands = []cli.Command{
		{
			Name:    "print",
			Aliases: []string{"p"},
			Usage:   "print data for this week",
			Action: func(c *cli.Context) error {
				var report map[string]int
				var week int
				normalized := c.Bool("normalized")
				if c.Bool("prev") {
					report, week = jobLogger.PrviousWeekSnapshot(normalized)
				} else {
					report, week = jobLogger.ThisWeekSnapshot(normalized)
				}
				fmt.Println(snapshotToString(report, week))
				return nil
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "normalized",
					Usage: "Normalize values. true means 100% is total. false means 100% is 40h. If you worked more then 40h this week and use false value, sum will be more then 100%",
				},
				cli.BoolFlag{
					Name: "prev",
					Usage: "Print for previous week instead of current",
				},
			},
		},
	}

	cliApp.Action = func(c *cli.Context) error {
		gui.NewQGuiApplication(len(os.Args), os.Args)

		bridge := NewQmlBride(nil)
		bridge.jobLogger = jobLogger
		bridge.clipboard = gui.QGuiApplication_Clipboard()
		bridge.showPeriodMs = uint64(configuration.AskPeriodMin * 60000)
		bridge.workingTimeYellowLimit = configuration.WorkingTimeYellowLimit
		bridge.workingTimeRedLimit = configuration.WorkingTimeRedLimit

		engine := qml.NewQQmlApplicationEngine(nil)
		engine.Load(core.NewQUrl3("qrc:/qml/application.qml", 0))
		ctxt := engine.RootContext()

		ctxt.SetContextProperty("bridge", bridge)
		ctxt.SetContextProperty("projectList", core.NewQStringListModel2(configuration.Projects, nil))

		gui.QGuiApplication_Exec()

		return nil
	}
	cliApp.Run(os.Args)
}
