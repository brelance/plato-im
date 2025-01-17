package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/jroimartin/gocui"
)

type User struct {
	Username string
}

var (
	currentUser User
	messageLog  []string
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Input area for typing messages
	if v, err := g.SetView("input", 0, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Input (Type message and press Enter)"
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
	}

	// Messages area for displaying messages
	if v, err := g.SetView("messages", 0, 0, maxX-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Messages"
		v.Wrap = true
	}

	// User area for showing the current username
	if v, err := g.SetView("user", 0, maxY-5, maxX-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "User Info"
		v.Wrap = true
		v.Clear()
		fmt.Fprintf(v, "Current User: %s", currentUser.Username)
	}

	return nil
}

func updateMessagesView(g *gocui.Gui) error {
	v, err := g.View("messages")
	if err != nil {
		return err
	}
	v.Clear()
	for _, msg := range messageLog {
		fmt.Fprintln(v, msg)
	}
	return nil
}

func sendMessage(g *gocui.Gui, v *gocui.View) error {
	messageInput := strings.TrimSpace(v.Buffer())
	if messageInput != "" {
		formattedMessage := fmt.Sprintf("%s: %s", currentUser.Username, messageInput)
		messageLog = append(messageLog, formattedMessage)
		updateMessagesView(g)
		v.Clear()
	}
	return nil
}

func setUsername(g *gocui.Gui, v *gocui.View) error {
	username := strings.TrimSpace(v.Buffer())
	if username != "" {
		currentUser.Username = username
		v.Clear()
		updateUserView(g)
	}
	return nil
}

func updateUserView(g *gocui.Gui) error {
	v, err := g.View("user")
	if err != nil {
		return err
	}
	v.Clear()
	fmt.Fprintf(v, "Current User: %s", currentUser.Username)
	return nil
}

func keybindings(g *gocui.Gui) error {
	// Keybinding to send a message
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, sendMessage); err != nil {
		return err
	}

	// Keybinding to set a new username (Ctrl+U)
	if err := g.SetKeybinding("", gocui.KeyCtrlU, gocui.ModNone, promptUsername); err != nil {
		return err
	}

	// Keybinding to quit the application (Ctrl+C)
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	return nil
}

func promptUsername(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("username", maxX/4, maxY/2-1, maxX*3/4, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Enter Username"
		v.Editable = true
		if _, err := g.SetCurrentView("username"); err != nil {
			return err
		}
	}

	// Keybinding to save the username when Enter is pressed
	if err := g.SetKeybinding("username", gocui.KeyEnter, gocui.ModNone, saveUsername); err != nil {
		return err
	}
	return nil
}

func saveUsername(g *gocui.Gui, v *gocui.View) error {
	setUsername(g, v)
	if err := g.DeleteView("username"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("input"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func IMUI() {
	currentUser = User{Username: "Guest"} // Default username

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalf("Failed to create GUI: %v", err)
	}
	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Fatalf("Keybindings setup error: %v", err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalf("Main loop error: %v", err)
	}
}
