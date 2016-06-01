package main

import (
	"fmt"
	"image"
	"os"

	"github.com/axet/desktop/go"
)

type SysTest struct {
	s *desktop.DesktopSysTray
}

func (m *SysTest) Click(mn *desktop.Menu) {
	fmt.Println("m", mn.Name)
}

func (m *SysTest) ClickBox(mn *desktop.Menu) {
	fmt.Println(mn.Name)
	mn.State = !mn.State
	m.s.Update()
}

const icon_png = `
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
`



func main() {
	m := SysTest{desktop.DesktopSysTrayNew()}

	file, err := os.Open("icon.png")
	if err != nil {
		panic(err)
	}
	icon, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	menu := []desktop.Menu{
		desktop.Menu{Icon: icon, Type: desktop.MenuItem, Enabled: true, Name: "test1", Action: m.Click},
		desktop.Menu{Type: desktop.MenuSeparator},
		desktop.Menu{Icon: icon, Type: desktop.MenuItem, Enabled: true, Name: "test2", Menu: []desktop.Menu{
			desktop.Menu{Type: desktop.MenuItem, Enabled: true, Name: "test21", Action: m.Click},
			desktop.Menu{Type: desktop.MenuItem, Enabled: true, Name: "test22", Action: m.Click},
		}},
		desktop.Menu{Type: desktop.MenuItem, Enabled: false, Name: "test3", Action: m.Click},
		desktop.Menu{Type: desktop.MenuCheckBox, Enabled: true, Name: "test4", State: true, Action: m.ClickBox},
		desktop.Menu{Type: desktop.MenuSeparator},
		desktop.Menu{Icon: icon, Type: desktop.MenuItem, Enabled: true, Name: "test5", Action: m.Click},
	}

	m.s.SetIcon(icon)
	m.s.SetTitle("go menu hoho!")
	m.s.SetMenu(menu)
	m.s.Show()

	desktop.Main()
}
