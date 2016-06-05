package main

import (
	log "github.com/Sirupsen/logrus"
//"github.com/jasonlvhit/gocron"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
	"github.com/jasonlvhit/gocron"
	"fmt"
	"os"
	"image"
	"github.com/axet/desktop/go"
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

func (m *App) CopyThisWeekToClipboard(mn *desktop.Menu) {
}
func (m *App) CopyPrevWeekToClipboard(mn *desktop.Menu) {
}

func (v *App) initWindow() {
	log.Info("Window creation...")
	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_POPUP)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	v.win = win
	v.win.SetTitle("What are you doing now?")
	v.win.Resize(300, 50)
	v.win.SetBorderWidth(10)
	v.win.SetPosition(gtk.WIN_POS_CENTER)
	v.win.Connect("destroy", func() {
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
	for _, project := range v.configuration.Projects {
		comboBox.AppendText(project)
	}

	comboBox.Connect("changed", v.on_change)

	comboBox.SetSizeRequest(270, 20)

	button, err := gtk.ButtonNew()
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	button.SetLabel("OK")
	button.Connect("clicked", v.on_button)

	grid2.Add(comboBox)
	grid2.Add(button)

	grid.Add(l)
	grid.Add(grid2)

	v.win.Add(grid)
	log.Info("Window created")
}

func (v *App) show() {
	log.Info("Show window")
	if !v.win.GetNoShowAll() {
		glib.IdleAdd(v.win.ShowAll)
	} else {
		v.jobLogger.AddForNow("===")
	}
}

func (v *App) on_change(text *gtk.ComboBoxText) {
	v.activeProject = text.GetActiveText()
	v.changeButtonState()
}

func (v *App) on_button() {
	v.jobLogger.AddForNow(v.activeProject)
	log.Info("Hide window")
	v.win.Hide()
	for key, value := range v.jobLogger.ThisWeekSnapshot() {
		fmt.Println("Key:", key, "Value:", value)
	}
}

func (v *App) changeButtonState() {

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
