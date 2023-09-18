package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

type LoginUser struct {
	list          []string
	width, height int
	g             *gocui.Gui
	viewName      string

	// selectedUser is the selected user when click Enter
	selectedUser string

	// pos indicates the current arrow position
	pos int
}

// NewLoginUser create a new login user UI
func NewLoginUser(listUser []string) (*LoginUser, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}
	g.Highlight = true
	g.Cursor = false
	g.SelFgColor = gocui.ColorGreen
	listUser = reverse(listUser)

	l := &LoginUser{
		list:     listUser,
		g:        g,
		viewName: "login_user_selector",
	}
	l.g.SetManagerFunc(func(gui *gocui.Gui) error {
		return l.registerView()
	})
	if err := l.registerKeyBind(); err != nil {
		return nil, err
	}

	l.width, l.height = g.Size()
	return l, nil
}

// getY return the initial start & end Y
func (l *LoginUser) getY() (yStart int, yEnd int) {
	textHeight := len(strings.Split(l.text(0), "\n"))
	paddingPercentage := (1 - float64(textHeight)/float64(l.height)) / 2
	yStart = int(float64(l.height)*paddingPercentage) - 1
	yEnd = yStart + textHeight
	return
}

// getX returns the initial start & end X
func (l *LoginUser) getX() (xStart, xEnd int) {
	xMax := 0
	for _, s := range l.list {
		if len(s) > xMax {
			xMax = len(s) + 10
		}
	}

	// when the xMax is lower than 35 characters
	// the rectangle popup is not precision
	// the button Cancel [CTRL+C] is cut
	// hence we need to set minimum X
	if xMax < 35 {
		xMax = 35
	}
	paddingPercentage := (1 - float64(xMax)/float64(l.width)) / 2
	xStart = int(float64(l.width)*paddingPercentage) - 1
	xEnd = xStart + xMax
	return
}

// registerView delete all previous view & keybinding
func (l *LoginUser) registerView() error {

	yStart, yEnd := l.getY()
	xStart, xEnd := l.getX()

	v, err := l.g.SetView(l.viewName, xStart, yStart, xEnd, yEnd)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	if err := l.write(v); err != nil {
		return err
	}

	if _, err := l.g.SetCurrentView(l.viewName); err != nil {
		return err
	}

	return nil
}

func (l *LoginUser) registerKeyBind() error {
	if err := l.g.SetKeybinding(l.viewName, gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := l.g.SetKeybinding(l.viewName, gocui.KeyEnter, gocui.ModNone, l.handleEnter); err != nil {
		return err
	}

	if err := l.g.SetKeybinding(l.viewName, gocui.KeyTab, gocui.ModNone, l.handleNav(func() {
		if l.pos < len(l.list)-1 {
			l.pos++
		}
	})); err != nil {
		return err
	}

	if err := l.g.SetKeybinding(l.viewName, gocui.KeyArrowUp, gocui.ModNone, l.handleNav(func() {
		if l.pos > 0 {
			l.pos--
		}
	})); err != nil {
		return err
	}
	if err := l.g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, l.handleNav(func() {
		if l.pos < len(l.list)-1 {
			l.pos++
		}
	})); err != nil {
		return err
	}
	return nil
}

// handleNav handle navigation arrow DOWN and UP
func (l *LoginUser) handleNav(c func()) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		c()
		return l.write(v)
	}
}

// write writes the text into view
func (l *LoginUser) write(v *gocui.View) error {
	v.Clear()
	_, err := v.Write([]byte(l.text(l.pos)))
	return err
}

// text return the text to be shown in the UI
func (l *LoginUser) text(pos int) string {
	var str bytes.Buffer
	str.WriteString("\n")
	str.WriteString("Select user to login")
	str.WriteString("\n\n")
	for i, s := range l.list {
		if i == pos {
			str.WriteString(fmt.Sprintf("\u001B[33;1m▶ %s\u001B[0m\n", s))
		} else {
			str.WriteString(fmt.Sprintf("  %s\n", s))
		}
	}
	str.WriteString("\n")
	str.WriteString("Yes [\u001B[32;1mEnter\u001B[0m]   Cancel [\u001B[31;1mCTRL+C\u001B[0m]")
	return prependTab(str.String())
}

func prependTab(text string) (res string) {
	for _, s := range strings.Split(text, "\n") {
		res += fmt.Sprintf("\n  %s ", s)
	}
	return
}

// handleEnter get the current list position then set to selected user
// then exit the UI
func (l *LoginUser) handleEnter(_ *gocui.Gui, _ *gocui.View) error {
	l.selectedUser = l.list[l.pos]
	return gocui.ErrQuit
}

// Run runs the UI and returns the selected user login
func (l *LoginUser) Run() (string, error) {
	defer l.g.Close()
	err := l.g.MainLoop()
	if err == gocui.ErrQuit {
		return l.selectedUser, nil
	}
	return l.selectedUser, err
}

// reverse will reverse the slice input to it (e.g. user list)
// @irf needs this since root is always on top
func reverse(s []string) []string {
	a := make([]string, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}
