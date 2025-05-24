package view

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app  *tview.Application = tview.NewApplication()
	flex *tview.Flex        = tview.NewFlex()

	passwordsList *tview.List       = tview.NewList()
	searchField   *tview.InputField = tview.NewInputField()
	statusText    *tview.TextView   = tview.NewTextView()

	onSearch func(string)                          = func(s string) {}
	onDone   func(text string, isNewPassword bool) = func(text string, isNewPassword bool) {}
	onDelete func(int)                             = func(i int) {}
)

func init() {
	flex.SetDirection(tview.FlexRow).
		AddItem(searchField, 1, 1, true).
		AddItem(passwordsList, 0, 1, false).
		AddItem(statusText, 1, 1, false)

	viewInit()
}

func Run() {
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatalf("Can't start user interface: %v", err)
	}
}

func SetItems(items []string) {
	passwordsList.Clear()
	for _, item := range items {
		passwordsList.AddItem(item, item, 0, nil)
	}
}

func SetOnSearchCallback(callback func(string)) {
	onSearch = callback
}

func SetOnDoneCallback(callback func(text string, isNewPassword bool)) {
	onDone = callback
}

func SetOnDeleteCallback(callback func(int)) {
	onDelete = callback
}

func SetStatusString(status string) {
	statusText.SetBackgroundColor(tcell.Color104)
	statusText.SetText(status)
}

func SetStatusErrorString(status string) {
	statusText.SetBackgroundColor(tcell.ColorDarkRed)
	statusText.SetText(status)
}

func RequestPassword(passwordChan chan string) {
	searchField.SetLabel("enter password:")
	searchField.SetText("")
	searchField.SetBackgroundColor(tcell.ColorDarkGreen)
	searchField.SetMaskCharacter('*')
	searchField.SetChangedFunc(nil)
	searchField.SetDoneFunc(func(key tcell.Key) {
		passwordChan <- searchField.GetText()
		close(passwordChan)
		viewInit()
	})
}

func Redraw() {
	app.Draw()
}

func viewInit() {
	passwordsList.ShowSecondaryText(false)

	searchField.SetLabel(":").
		SetText("").
		SetMaskCharacter(0).
		SetDoneFunc(search).
		SetChangedFunc(func(text string) {
			onSearch(text)
		})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlN, tcell.KeyDown:
			passwordsList.SetCurrentItem(passwordsList.GetCurrentItem() + 1)

		case tcell.KeyCtrlP, tcell.KeyUp:
			passwordsList.SetCurrentItem(passwordsList.GetCurrentItem() - 1)

		case tcell.KeyCtrlD:
			if isPasswordSelected() {
				onDelete(passwordsList.GetCurrentItem())
			}

		case tcell.KeyTab:
			if !isPasswordSelected() {
				return nil
			}
			currentText := searchField.GetText()
			_, firstPassword := passwordsList.GetItemText(0)
			first, _, found := strings.Cut(firstPassword[len(currentText):], "/")
			if found {
				first = first + "/"
			}
			searchField.SetText(currentText + first)
			return nil

		case tcell.KeyCtrlQ:
			app.Stop()
		}
		return event
	})
}

func search(key tcell.Key) {
	if !isPasswordSelected() {
		onDone(searchField.GetText(), true)
	} else {
		_, secondaryText := passwordsList.GetItemText(passwordsList.GetCurrentItem())
		onDone(secondaryText, false)
	}
}

func isPasswordSelected() bool {
	return passwordsList.GetItemCount() != 0
}
