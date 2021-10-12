package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
	"github.com/qeesung/image2ascii/convert"
)

type windowLayout struct {
	views.BoxLayout
	width  int
	height int
}

type textField struct {
	views.TextArea
	symbols   []rune
	hideInput bool
}

type contentArea struct {
	views.Text
}

var window = &windowLayout{}
var field = &textField{}
var content = &contentArea{}
var app = &views.Application{}
var art *views.Text
var spacer2 *views.Text
var spacer3 *views.Text

var sbuilder strings.Builder
var img image.Image
var screen tcell.Screen

func (window *windowLayout) HandleEvent(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventInterrupt:
		spacer3.SetText(fmt.Sprint(event.When().Clock()))
		app.Update()

	case *tcell.EventKey:
		if event.Key() == tcell.KeyEscape {
			app.Quit()
			return true
		}
		if event.Key() == tcell.KeyTab {
			field.hideInput = !field.hideInput
			field.HideCursor(field.hideInput)
			field.EnableCursor(!field.hideInput)
			field.MakeCursorVisible()
		}
	}
	return window.BoxLayout.HandleEvent(event)
}

func (window *windowLayout) checkOrientation() {
	width, height := screen.Size()
	if window.width != width || window.height != height {
		if width > 2*height {
			window.SetOrientation(views.Horizontal)
			spacer2.SetText("   ")
		} else {
			window.SetOrientation(views.Vertical)
			spacer2.SetText("")
		}
		window.width = width
		window.height = height
		art.SetText(refitArt())
	}
}

func (field *textField) HandleEvent(event tcell.Event) bool {
	if field.hideInput {
		return true
	}
	switch event := event.(type) {
	case *tcell.EventKey:
		posX, _, _, _ := field.GetModel().GetCursor()
		switch event.Key() {

		case tcell.KeyEnter:
			field.Clear()
			field.hideInput = !field.hideInput
			field.HideCursor(field.hideInput)
			field.EnableCursor(!field.hideInput)

		case tcell.KeyBackspace2:
			if posX > 0 {
				posX--
				field.symbols[posX] = 0
				field.symbols = append(field.symbols[:posX],
					field.symbols[posX+1:]...)
			}
			field.SetContent(string(field.symbols))
			field.SetCursorX(posX)

		case tcell.KeyDelete:
			if posX < len(field.symbols)-1 {
				field.symbols[posX] = 0
				field.symbols = append(field.symbols[:posX],
					field.symbols[posX+1:]...)
				posX++
			}
			field.SetContent(string(field.symbols))

		case tcell.KeyRune:
			field.symbols = append(field.symbols, 0)
			copy(field.symbols[posX+1:], field.symbols[posX:])
			field.symbols[posX] = event.Rune()
			field.SetContent(string(field.symbols))
			field.SetCursorX(posX + 1)
		}
	}
	return field.TextArea.HandleEvent(event)
}

func (field *textField) getText() string {
	for i, r := range field.symbols {
		// trailing space doesn't need to be in actual input
		if i == len(field.symbols)-1 {
			break
		}
		fmt.Fprint(&sbuilder, string(r))
	}
	defer sbuilder.Reset()

	return sbuilder.String()
}

func (field *textField) Clear() {
	field.SetContent(" ")
	field.symbols = make([]rune, 1)
	field.symbols[0] = ' '
	field.SetCursorX(0)
}

func (content *contentArea) HandleEvent(event tcell.Event) bool {
	switch event := event.(type) {
	case *views.EventWidgetResize:
		window.checkOrientation()
	case *tcell.EventKey:
		if event.Key() == tcell.KeyEnter {
			content.SetText(field.getText())
		}
	}
	return content.Text.HandleEvent(event)
}

func refitArt() string {
	options := convert.Options{
		Ratio:           1.0,
		FixedWidth:      -1,
		FixedHeight:     -1,
		FitScreen:       true,
		StretchedScreen: false,
		Colored:         false,
		Reversed:        false,
	}
	return convert.NewImageConverter().Image2ASCIIString(img, &options)
}

func checkFatalError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	file, err := os.Open("./gopher.png")
	checkFatalError(err)

	img, err = png.Decode(file)
	checkFatalError(err)
	file.Close()

	margin := "   "

	spacer1 := views.NewText()
	spacer1.SetText(margin)
	spacer1.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorChocolate).
		Background(tcell.ColorSkyblue))

	art = views.NewText()
	art.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke).
		Background(tcell.ColorTomato))

	spacer2 = views.NewText()
	spacer2.SetText(margin)
	spacer2.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorChocolate).
		Background(tcell.ColorSkyblue))

	contentBox := views.NewBoxLayout(views.Vertical)

	spacer3 = views.NewText()
	spacer3.SetText(margin)
	spacer3.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorChocolate).
		Background(tcell.ColorSkyblue))

	content.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).
		Background(tcell.ColorLightGoldenrodYellow))

	message := views.NewText()
	message.SetText("error message goes here")
	message.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).
		Background(tcell.ColorPaleGoldenrod))

	field.hideInput = true
	field.EnableCursor(!field.hideInput)
	field.HideCursor(field.hideInput)
	field.Clear()
	field.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorDarkSlateGray).
		Background(tcell.ColorYellowGreen))

	spacer4 := views.NewText()
	spacer4.SetText(margin)
	spacer4.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorChocolate).
		Background(tcell.ColorSkyblue))

	window.AddWidget(spacer1, 0)
	window.AddWidget(art, 0.0)
	window.AddWidget(spacer2, 0)
	contentBox.AddWidget(spacer3, 0)
	contentBox.AddWidget(content, 1.0)
	contentBox.AddWidget(message, 0)
	contentBox.AddWidget(field, 0)
	window.AddWidget(contentBox, 1.0)
	window.AddWidget(spacer4, 0)

	screen, err = tcell.NewScreen()
	checkFatalError(err)

	app.SetScreen(screen)
	app.SetRootWidget(window)

	app.PostFunc(func() {
		window.width, window.height = screen.Size()
	})
	app.PostFunc(func() {
		go func() {
			for {
				time.Sleep(time.Second / 2)
				window.HandleEvent(tcell.NewEventInterrupt(nil))
			}
		}()
	})
	err = app.Run()
	checkFatalError(err)
}
