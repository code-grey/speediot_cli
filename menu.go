package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)



func showMainMenu(s tcell.Screen, style tcell.Style) (string, string) {
	asciiArt := []string{
		`																`,
		`                         _ _       _                   _ _ 	`,
		`                        | (_)     | |                 | (_)	`,
		` ___ _ __   ___  ___  __| |_  ___ | |_             ___| |_ 	`,
		`/ __| '_ \ / _ \/ _ \/ _' | |/ _ \| __|           / __| | |	`,
		`\__ \ |_) |  __/  __/ (_| | | (_) | |_           | (__| | |	`,
		`|___/ .__/ \___|\___|\__,_|_|\___/ \__|           \___|_|_|	`,
		`    | |                                  ______            	`,
		`    |_|                                 |______|           	`,
	}

	options := []string{
		"1. Easy",
		"2. Medium",
		"3. Hard",
		"4. Dynamic (from DB)",
	}
	selected := 0

	for {
		s.Clear()
		yOffset := 0
		for _, line := range asciiArt {
			printString(s, 0, yOffset, line, style)
			yOffset++
		}

		printString(s, 0, yOffset+1, "Select Difficulty:", style)
		for i, opt := range options {
			currentStyle := style
			if i == selected {
				currentStyle = style.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
			}
			printString(s, 0, yOffset+3+i, opt, currentStyle)
		}

		// Display Scoreboard
		scores, err := GetTopScoresFromDB()
		if err != nil {
			log.Printf("Error getting scores: %v", err)
		} else {
			printString(s, 0, yOffset+3+len(options)+2, "--- Scoreboard (Top 10) ---", style)
			if len(scores) == 0 {
				printString(s, 0, yOffset+3+len(options)+3, "No scores yet. Play a test!", style)
			} else {
				for i, score := range scores {
					scoreLine := fmt.Sprintf("%d. %s - Score: %.2f (WPM: %.2f, Accuracy: %.2f%%)", i+1, score.Username, score.CalculatedScore, score.WPM, score.Accuracy)
					printString(s, 0, yOffset+3+len(options)+3+i, scoreLine, style)
				}
			}
		}

		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyUp {
				selected = (selected - 1 + len(options)) % len(options)
			} else if ev.Key() == tcell.KeyDown {
				selected = (selected + 1) % len(options)
			} else if ev.Key() == tcell.KeyEnter {
				switch selected {
				case 0:
					return easyText, getUsername(s, style)
				case 1:
					return mediumText, getUsername(s, style)
				case 2:
					return hardText, getUsername(s, style)
				case 3:
					text, err := GetRandomTextFromDB()
					if err != nil {
						log.Printf("Error getting dynamic text: %v", err)
						return easyText, getUsername(s, style) // Fallback to easy text on error
					}
					return text, getUsername(s, style)
				}
			} else if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				os.Exit(0)
			}
		}
	}
}

func askToRestart(s tcell.Screen, style tcell.Style, wpm, accuracy float64) int {
	options := []string{
		"1. Play again",
		"2. Go to Main Menu",
		"3. Exit",
	}
	selected := 0

	for {
		s.Clear()
		printString(s, 0, 0, fmt.Sprintf("Your Score: WPM: %.2f, Accuracy: %.2f%%", wpm, accuracy), style)
		printString(s, 0, 2, "What would you like to do?", style)

		for i, opt := range options {
			currentStyle := style
			if i == selected {
				currentStyle = style.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
			}
			printString(s, 0, 4+i, opt, currentStyle)
		}
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyUp {
				selected = (selected - 1 + len(options)) % len(options)
			} else if ev.Key() == tcell.KeyDown {
				selected = (selected + 1) % len(options)
			} else if ev.Key() == tcell.KeyEnter {
				switch selected {
				case 0: // Play again
					return ActionPlayAgain
				case 1: // Go to Main Menu
					return ActionGoToMainMenu
				case 2: // Exit
					os.Exit(0)
				}
			} else if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				os.Exit(0)
			}
		}
	}
}

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