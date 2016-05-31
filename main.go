package main

import (
	log "github.com/Sirupsen/logrus"
//"github.com/jasonlvhit/gocron"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/glib"
	"github.com/jasonlvhit/gocron"
	"fmt"
)

const (
	OTHER = "other"
	PROJECT1 = "project 1"
	PROJECT2 = "project 2"
)

type AskBox struct {
	activeProject string ""
	jobLogger     JobLogger
	win           *gtk.Window
}

func (v *AskBox) initWindow() {
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
	comboBox.AppendText(OTHER)
	comboBox.AppendText(PROJECT1)
	comboBox.AppendText(PROJECT2)

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

func (v *AskBox) show() {
	log.Info("Show window")
	if !v.win.GetNoShowAll() {
		glib.IdleAdd(v.win.ShowAll)
	} else {
		v.jobLogger.AddForNow("===")
	}
}

func (v *AskBox) on_change(text *gtk.ComboBoxText) {
	v.activeProject = text.GetActiveText()
	v.changeButtonState()
}

func (v *AskBox) on_button() {
	v.jobLogger.AddForNow(v.activeProject)
	log.Info("Hide window")
	v.win.Hide()
	for key, value := range  v.jobLogger.ThisWeekSnapshot() {
		fmt.Println("Key:", key, "Value:", value)
	}
}

func (v *AskBox) changeButtonState() {

}

func mkgrid(orientation gtk.Orientation) *gtk.Grid {
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(orientation)
	return grid
}

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp:true, TimestampFormat:"2006-01-02 15:04:05 -0700"})
	log.Info("Starting...")

	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	askBox := AskBox{
		jobLogger: FileJobLogger{
			Basedir:"/tmp",
		},
	}

	askBox.initWindow()

	s := gocron.NewScheduler()
	periodMin := 30
	log.Info(fmt.Sprintf("Scheduling cron task for every %d minutes", periodMin))
	s.Every(uint64(periodMin)).Minutes().Do(askBox.show)
	s.Start()

	askBox.show()
	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}
