package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
)

const (
	listViewWidth     = 40 // Ширина представления списка
	inputOutputHeight = 3  // Высота представлений "input" и "output"
)

func runGUI(processes ProcessMap, channels ProcChannels) error {
	clientGUI, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logger.Println("Failed to create a GUI object:", err)
		return err
	}
	defer clientGUI.Close()

	clientGUI.Cursor = true
	clientGUI.SetManagerFunc(layout) // Устанавливаем функцию-обработчик для размещения представлений, чтобы в случае обновления заново перерисовалось

	setKeyBindings(&processes, channels, clientGUI) // Устанавливаем обработчики клавиш

	terminalWidth, terminalHeight := clientGUI.Size()

	createViews(clientGUI, terminalWidth, terminalHeight)
	go

	return nil
}

func createViews(gui *gocui.Gui, width int, height int) interface{} {

	terminalWidth, terminalHeight := gui.Size()

	{
		// Создаем представление "list" для списка процессов
		listView, err := gui.SetView("list", 0, 0, listViewWidth, terminalHeight-inputOutputHeight-1)
		if err != nil && err != gocui.ErrUnknownView {
			return errors.Wrap(err, "Failed to create list view")
		}
		listView.Title = "List"
		listView.FgColor = gocui.ColorCyan // TOOD change colour

		// Создаем представление "output" для вывода результатов выполнения команд
	}

	{
		outputView, err := gui.SetView("output", listViewWidth+1, 0, terminalWidth-1, terminalHeight-inputOutputHeight-1)
		if err != nil && err != gocui.ErrUnknownView {
			return errors.Wrap(err, "Failed to create output view")
		}
		outputView.Title = "Output"
		outputView.FgColor = gocui.ColorDefault // TOOD change colour

		outputView.Autoscroll = true // Включаем автопрокрутку
		_, err = fmt.Fprintf(outputView, "Ctrl-C to quit:\n")
		if err != nil {
			return errors.Wrap(err, "Failed to write to output view")
		}
	}
	{
		// Создаем представление "input" для ввода команд
		inputView, err := gui.SetView("input", listViewWidth+1, terminalHeight-inputOutputHeight, terminalWidth-1, terminalHeight-1)
		if err != nil && err != gocui.ErrUnknownView {
			return errors.Wrap(err, "Failed to create input view")
		}
		inputView.Title = "Input"
		inputView.FgColor = gocui.ColorBlue // TOOD change colour
	}
	return nil
}

func setKeyBindings(p *ProcessMap, channels ProcChannels, gui *gocui.Gui) {
	// Устанавливаем обработчик для комбинации клавиш Ctrl+C
	err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
		func(gui *gocui.Gui, view *gocui.View) error {
			return quit(gui, view)
		})
	if err != nil {
		errors.Wrap(err, "Failed to set key binding")
		return
	}

	// Устанавливаем обработчик для клавиши Enter
	err = gui.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, wrap(p, channels))
	if err != nil {
		errors.Wrap(err, "Failed to set key binding")
	}
}

func quit(gui *gocui.Gui, view *gocui.View) error {
	return gocui.ErrQuit
}

func layout(gui *gocui.Gui) error {
	terminalWidth, terminalHeight := gui.Size()

	_, err := gui.SetView("list", 0, 0, listViewWidth, terminalHeight-inputOutputHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		errors.Wrap(err, "Failed to create list view")
	}
	_, err = gui.SetView("output", listViewWidth+1, 0, terminalWidth-1, terminalHeight-inputOutputHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		errors.Wrap(err, "Failed to create output view")
	}
	_, err = gui.SetView("input", 0, terminalHeight-inputOutputHeight, terminalWidth-1, terminalHeight-1)
	if err != nil && err != gocui.ErrUnknownView {
		errors.Wrap(err, "Failed to create input view")
	}
	return nil
}
