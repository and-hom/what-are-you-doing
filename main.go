package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
	"github.com/jasonlvhit/gocron"
	"fmt"
	"image"
	"bytes"
	"github.com/axet/desktop/go"
	"github.com/atotto/clipboard"
	"time"
	"os"
)

func mkgrid(orientation gtk.Orientation) *gtk.Grid {
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(orientation)
	return grid
}

func assertBtnCreated(err error) {
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
}

type App struct {
	configuration Configuration
	systray       *desktop.DesktopSysTray
	jobLogger     JobLogger
	activeProject string ""
	win           *gtk.Window
}

func (app *App) toClipboardText(text string) {
	if err := clipboard.WriteAll(text); err != nil {
		panic(err)
	}
}
func (app *App) toClipboard(report map[string]int, week int) {
	var buffer bytes.Buffer

	local, _ := time.LoadLocation("Local")
	monday := firstDayOfISOWeek(time.Now().Year(), week, local)
	friday := monday.Add((5 * 24 -1) * time.Hour)
	buffer.WriteString(fmt.Sprintf("====%s - %s\n", monday.Format("2006-01-02"), friday.Format("2006-01-02")))
	for k, v := range report {
		buffer.WriteString(fmt.Sprintf("**%s** - %d%%\n", k, v))
	}

	app.toClipboardText(buffer.String())
}

func (app *App) CopyThisWeekToClipboard(mn *desktop.Menu) {
	report, week := app.jobLogger.ThisWeekSnapshot(app.configuration.PercentageMode)
	app.toClipboard(report, week)
}
func (app *App) CopyPrevWeekToClipboard(mn *desktop.Menu) {
	report, week := app.jobLogger.PrviousWeekSnapshot(app.configuration.PercentageMode)
	app.toClipboard(report, week)
}
func (app *App) copyCurrentWeekNumberToClipboard(mn *desktop.Menu) {
	app.toClipboardText(fmt.Sprintf("%02d", thisWeek()))
}
func (app *App) exit(mn *desktop.Menu) {
	os.Exit(0)
}
func (app *App) initWindow() {
	log.Info("Window creation...")
	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_POPUP)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	app.win = win
	app.win.SetTitle("What are you doing now?")
	app.win.Resize(200, 50 + len(app.configuration.Projects) * 20)
	app.win.SetBorderWidth(10)
	app.win.SetPosition(gtk.WIN_POS_CENTER)
	app.win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	grid := mkgrid(gtk.ORIENTATION_VERTICAL)

	l, err := gtk.LabelNew("What are you doing now?")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	l.SetMarginBottom(5)
	l.SetMarginEnd(50)

	skipBtn, err := gtk.ButtonNew()
	assertBtnCreated(err)
	skipBtn.SetLabel("Skip")
	skipBtn.Connect("clicked", app.on_skip)

	header := mkgrid(gtk.ORIENTATION_HORIZONTAL)
	header.Add(l)
	header.Add(skipBtn)

	grid2 := mkgrid(gtk.ORIENTATION_HORIZONTAL)

	okBtn, err := gtk.ButtonNew()
	assertBtnCreated(err)
	okBtn.SetLabel("OK")
	okBtn.Connect("clicked", app.on_button)

	grid2.Add(okBtn)

	grid.Add(header)
	grid.Add(grid2)

	var group *glib.SList = nil
	for i, project := range app.configuration.Projects {
		if (i == 0) {
			rb, _ := gtk.RadioButtonNewWithLabel(glib.WrapSList(0), project)
			gr, _ := rb.GetGroup()
			group = gr
			rb.Connect("clicked", app.on_change)
			grid.Add(rb)
		} else {
			rb, _ := gtk.RadioButtonNewWithLabel(group, project)
			gr, _ := rb.GetGroup()
			group = gr

			rb.Connect("clicked", app.on_change)
			grid.Add(rb)
		}
	}

	app.win.Add(grid)
	log.Info("Window created")
}

func (app *App) show() {
	log.Info("Show window")
	if !app.win.GetNoShowAll() {
		glib.IdleAdd(app.win.ShowAll)
	} else {
		app.jobLogger.AddForNow("===")
	}
}

func (app *App) on_change(text *gtk.RadioButton) {
	app.activeProject, _ = text.GetLabel()
	app.changeButtonState()
}

func (app *App) on_button() {
	app.jobLogger.AddForNow(app.activeProject)
	log.Infof("Hide window - selected %s", app.activeProject)
	app.win.Hide()
	report, _ := app.jobLogger.ThisWeekSnapshot(app.configuration.PercentageMode)
	for key, value := range report {
		fmt.Println("Key:", key, "Value:", value)
	}
}

func (app *App) on_skip() {
	app.jobLogger.AddForNow(app.activeProject)
	log.Info("Hide window - skip")
	app.win.Hide()
}

func (app *App) on_mode_week(mn *desktop.Menu) {
	app.configuration.PercentageMode = OfWeek
	app.systray.SetMenu(app.trayMenu(OfWeek))
	app.systray.Update()
}

func (app *App) on_mode_total(mn *desktop.Menu) {
	app.configuration.PercentageMode = OfTotal
	app.systray.SetMenu(app.trayMenu(OfTotal))
	app.systray.Update()

}

func (app *App) changeButtonState() {

}

func loadIcon() image.Image {
	imgData, err := Asset("icon.png")
	if err != nil {
		panic(err)
	}
	icon, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		panic(err)
	}
	return icon
}

func main() {

	log.SetFormatter(&log.TextFormatter{FullTimestamp:true, TimestampFormat:"2006-01-02 15:04:05 -0700"})
	log.Info("Starting...")

	configuration := load("")
	log.Infof("Loaded configuration %+v", configuration)

	app := App{
		systray:desktop.DesktopSysTrayNew(),
		jobLogger:FileJobLogger{
			Basedir:configuration.LogPath,
			AskPerHour:60 / configuration.AskPeriodMin,
		},
		configuration: configuration,
	}
	icon := loadIcon()

	app.systray.SetIcon(icon)
	app.systray.SetTitle("What are you doing now?")
	app.systray.SetMenu(app.trayMenu(configuration.PercentageMode))
	app.systray.Show()

	app.initWindow()

	s := gocron.NewScheduler()
	log.Info(fmt.Sprintf("Scheduling cron task for every %d minutes", app.configuration.AskPeriodMin))
	s.Every(uint64(app.configuration.AskPeriodMin)).Minutes().Do(app.show)
	s.Start()

	app.show()
	desktop.Main()
}

func (app *App) trayMenu(currentPercentageMode PercentageMode) []desktop.Menu {
	percentageMenus := []desktop.Menu{
		desktop.Menu{
			Type:desktop.MenuCheckBox,
			Enabled:true,
			Name:"Of week (40h)",
			State:currentPercentageMode == OfWeek,
			Action:app.on_mode_week,
		},
		desktop.Menu{
			Type:desktop.MenuCheckBox,
			Enabled:true,
			Name:"Of total spent time",
			State:currentPercentageMode == OfTotal,
			Action:app.on_mode_total,
		},
	}
	return []desktop.Menu{
		desktop.Menu{
			Type:desktop.MenuItem,
			Enabled:true,
			Name:fmt.Sprintf("w%02d", thisWeek()),
			Action:app.copyCurrentWeekNumberToClipboard,
		},
		desktop.Menu{
			Type: desktop.MenuItem,
			Enabled: true,
			Name: "Copy this week report to clipboard",
			Action: app.CopyThisWeekToClipboard,
		},
		desktop.Menu{
			Type: desktop.MenuItem,
			Enabled: true,
			Name: "Copy prev week report to clipboard",
			Action: app.CopyPrevWeekToClipboard,
		},
		desktop.Menu{
			Type:desktop.MenuItem,
			Enabled:true,
			Name:"Percent calculation mode",
			Menu:percentageMenus,
		},
		desktop.Menu{
			Type:desktop.MenuItem,
			Enabled:true,
			Name:"Exit",
			Action:app.exit,
		},
	}
}
