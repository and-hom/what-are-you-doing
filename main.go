package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
	"github.com/jasonlvhit/gocron"
	"fmt"
	"os"
	"image"
	"bytes"
	"github.com/axet/desktop/go"
	"github.com/atotto/clipboard"
	"time"
)

func mkgrid(orientation gtk.Orientation) *gtk.Grid {
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(orientation)
	return grid
}

type App struct {
	configuration Configuration
	systray       *desktop.DesktopSysTray
	jobLogger     JobLogger
	activeProject string ""
	win           *gtk.Window
}

func (app *App) toClipboard(report map[string]int, week int) {
	var buffer bytes.Buffer

	local, _ := time.LoadLocation("Local")
	monday := firstDayOfISOWeek(time.Now().Year(), week, local)
	friday := monday.Add(5 * 24 * time.Hour)
	buffer.WriteString(fmt.Sprintf("%s - %s\n", monday.Format("2006-01-02"), friday.Format("2006-01-02")))
	for k, v := range report {
		buffer.WriteString(fmt.Sprintf("**%s** - %d%%\n", k, v))
	}

	if err := clipboard.WriteAll(buffer.String()); err != nil {
		panic(err)
	}
}

func (app *App) CopyThisWeekToClipboard(mn *desktop.Menu) {
	report, week := app.jobLogger.ThisWeekSnapshot()
	app.toClipboard(report, week)
}
func (app *App) CopyPrevWeekToClipboard(mn *desktop.Menu) {
	report, week := app.jobLogger.PrviousWeekSnapshot()
	app.toClipboard(report, week)
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
	app.win.Resize(300, 50)
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

	grid2 := mkgrid(gtk.ORIENTATION_HORIZONTAL)

	comboBox, err := gtk.ComboBoxTextNew()
	if err != nil {
		log.Fatal("Unable to create combo box:", err)
	}
	comboBox.AppendText("Other")
	for _, project := range app.configuration.Projects {
		comboBox.AppendText(project)
	}

	comboBox.Connect("changed", app.on_change)

	comboBox.SetSizeRequest(270, 20)

	button, err := gtk.ButtonNew()
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	button.SetLabel("OK")
	button.Connect("clicked", app.on_button)

	grid2.Add(comboBox)
	grid2.Add(button)

	grid.Add(l)
	grid.Add(grid2)

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

func (app *App) on_change(text *gtk.ComboBoxText) {
	app.activeProject = text.GetActiveText()
	app.changeButtonState()
}

func (app *App) on_button() {
	app.jobLogger.AddForNow(app.activeProject)
	log.Info("Hide window")
	app.win.Hide()
	report, _ := app.jobLogger.ThisWeekSnapshot()
	for key, value := range report {
		fmt.Println("Key:", key, "Value:", value)
	}
}

func (app *App) changeButtonState() {

}

func loadIcon(app App) image.Image {
	file, err := os.Open("icon.png")
	if err != nil {
		panic(err)
	}
	icon, _, err := image.Decode(file)
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
		},
		configuration: configuration,
	}
	icon := loadIcon(app)
	menu := []desktop.Menu{
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
	}

	app.systray.SetIcon(icon)
	app.systray.SetTitle("What are you doing now?")
	app.systray.SetMenu(menu)
	app.systray.Show()

	app.initWindow()

	s := gocron.NewScheduler()
	log.Info(fmt.Sprintf("Scheduling cron task for every %d minutes", app.configuration.AskPeriodMin))
	s.Every(uint64(app.configuration.AskPeriodMin)).Minutes().Do(app.show)
	s.Start()

	app.show()
	desktop.Main()
}
