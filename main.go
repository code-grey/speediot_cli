package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)




func getUsername(s tcell.Screen, style tcell.Style) string {
	username := ""
	prompt := "Enter your username (default: Guest): "

	for {
		s.Clear()
		printString(s, 0, 0, prompt+username, style)
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEnter {
				if username == "" {
					return "Guest"
				}
				return username
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(username) > 0 {
					username = username[:len(username)-1]
				}
			} else if ev.Key() == tcell.KeyRune {
				username += string(ev.Rune())
			}
		}
	}
}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err = s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	correctStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
	incorrectStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed)
	currentStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)

	s.SetStyle(defStyle)
	s.Clear()

	username := getUsername(s, defStyle)

	for {
		selectedText := showMainMenu(s, defStyle)
		wpm, accuracy := runTest(s, selectedText, defStyle, correctStyle, incorrectStyle, currentStyle, username)
		action := askToRestart(s, defStyle, wpm, accuracy)
		if action == ActionExit {
			break
		} else if action == ActionGoToMainMenu {
			continue
		}
		// If ActionPlayAgain, the loop continues naturally
	}

	s.Fini()
}






