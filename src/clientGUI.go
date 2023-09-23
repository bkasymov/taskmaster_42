package main

import (
	"github.com/jroimartin/gocui"
)

const (
	lw = 40

	ih = 4
)

func runGUI(processes ProcessMap, channels ProcChannels) error {
	clientGUI, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logger.Println("Failed to create a GUI object:", err)
		return err
	}
	defer clientGUI.Close()

	clientGUI.Cursor = true

	clientGUI.SetManagerFunc(layout)
	setKeyBindings(&processes, channels, clientGUI)
	return nil
}

func layout(gui *gocui.Gui) error {

}

//
// "" mean all view windows in the GUI
// ModNone means no modifier key is pressed (Ctrl, Alt, etc.)
// If client presses Ctrl+C, then quit the GUI
// we write what key is pressed, and what function is called
//
func setKeyBindings(processes *ProcessMap, channels ProcChannels, gui interface{}) {

	err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
		func(gui *gocui.Gui, view *gocui.View) error {
			return quit(gui, view)
		})
	if err != nil {
		logger.Println("Cannot bind the quit key", err)
		return
	}
	err = gui.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, wrap(processes, channels))
	if err != nil {
		logger.Println("Cannot bind the enter key:", err)
	}
}

func wrap(processMap *ProcessMap, processChannels ProcChans) func(gui *gocui.Gui, view *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		inputView, err := gui.View("input")
		if err != nil {
			logger.Println("Cannot get input view:", err)
			return err
		}

		outputView, err := gui.View("output")
		if err != nil {
			logger.Println("Cannot get output view:", err)
			return err
		}

		inputView.Rewind()
		commandLine := inputView.Buffer()
		getCommand(commandLine, processMap, processChannels, outputView)

		inputView.Clear()

		err = inputView.SetCursor(0, 0)
		if err != nil {
			logger.Println("Failed to set cursor:", err)
		}
		return err
	}
}

//
//func getCommand(line string, processMap *ProcessMap, channels interface{}, view *gocui.View) {
//
//}
//
//func quit(gui *gocui.Gui, view *gocui.View) error {
//	return gocui.ErrQuit
//}
