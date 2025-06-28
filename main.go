package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)






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

	

	for {
		selectedText, username, difficulty := showMainMenu(s, defStyle)
		wpm, accuracy := runTest(s, selectedText, defStyle, correctStyle, incorrectStyle, currentStyle, username, difficulty)
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






