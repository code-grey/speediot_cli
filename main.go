package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"
	"log"

	"github.com/gdamore/tcell/v2"
	_ "github.com/mattn/go-sqlite3"
)

const (
	easyText = "The quick brown fox jumps over the lazy dog. This is a simple typing test. Practice makes perfect."
	mediumText = "A journey of a thousand miles begins with a single step. The early bird catches the worm. All that glitters is not gold."
	hardText = "The 1st rule of 2025 is: Never give up! @#$! Success often comes after a string of failures. Are you ready for the challenge? (Press ESC to exit)"

	ActionPlayAgain    = 0
	ActionGoToMainMenu = 1
	ActionExit         = 2
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
		selectedText := showMainMenu(s, defStyle)
		wpm, accuracy := runTest(s, selectedText, defStyle, correctStyle, incorrectStyle, currentStyle)
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

func runTest(s tcell.Screen, textToType string, defStyle, correctStyle, incorrectStyle, currentStyle tcell.Style) (float64, float64) {
	rand.Seed(time.Now().UnixNano())
	words := strings.Fields(textToType)
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
	testText := strings.Join(words, " ")

	typedText := make([]rune, 0, len(testText))
	cursorPos := 0
	errors := 0 // Reintroduce errors counter
	startTime := time.Time{}
	testStarted := false

	for {
		s.Clear()
		printText(s, testText, typedText, cursorPos, correctStyle, incorrectStyle, currentStyle, defStyle)
		wpm, accuracy := drawStats(s, startTime, testStarted, typedText, testText, defStyle, cursorPos, errors)
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return 0, 0 // Corrected: return values for escape
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if cursorPos > 0 {
					_, size := utf8.DecodeLastRuneInString(string(typedText[:cursorPos]))
					typedText = typedText[:cursorPos-size]
					cursorPos -= size
				}
			} else if ev.Key() == tcell.KeyRune {
				if !testStarted {
					startTime = time.Now()
					testStarted = true
				}
				typedText = append(typedText, ev.Rune())
				if cursorPos < len(testText) {
					if ev.Rune() != rune(testText[cursorPos]) {
						errors++ // Increment errors only on incorrect input
					}
				}
				cursorPos += utf8.RuneLen(ev.Rune())
			}

			if cursorPos >= len(testText) && testStarted {
				// Test finished
				s.Clear()
				printText(s, testText, typedText, cursorPos, correctStyle, incorrectStyle, currentStyle, defStyle)
				wpm, accuracy = drawStats(s, startTime, testStarted, typedText, testText, defStyle, cursorPos, errors)
				s.Show()

				if testStarted {
					saveScore(wpm, accuracy)
				}
				return wpm, accuracy
			}
		}
	}
	return 0, 0 // Should not be reached, but required for function signature
}

func showMainMenu(s tcell.Screen, style tcell.Style) string {
	asciiArt := []string{
		`  _______ _ __   __ _           _       `,
		` /  ___  (_)  _ \ / _(_)         | |      `,
		`|  /  /  |_| |_) | |_ _ _ __ ___ | | ___  `,
		`| |   | | |  _ <|  _| | '_ ` + "`" + ` _ \| |/ _ \ `,
		`|  \__/  | | |_) | | | | | | | | | |  __/ `,
		` \_______|_|____/|_| |_|_| |_| |_|_|\___|`,
		`                                         `,
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
		scores, err := getTopScoresFromDB()
		if err != nil {
			log.Printf("Error getting scores: %v", err)
		} else {
			printString(s, 0, yOffset+3+len(options)+2, "--- Scoreboard (Top 10) ---", style)
			if len(scores) == 0 {
				printString(s, 0, yOffset+3+len(options)+3, "No scores yet. Play a test!", style)
			} else {
				for i, score := range scores {
					scoreLine := fmt.Sprintf("%d. WPM: %.2f, Accuracy: %.2f%%", i+1, score.WPM, score.Accuracy)
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
					return easyText
				case 1:
					return mediumText
				case 2:
					return hardText
				case 3:
					text, err := getRandomTextFromDB()
					if err != nil {
						log.Printf("Error getting dynamic text: %v", err)
						return easyText // Fallback to easy text on error
					}
					return text
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

func printText(s tcell.Screen, testText string, typedText []rune, cursorPos int, correctStyle, incorrectStyle, currentStyle, defStyle tcell.Style) {
	w, _ := s.Size()
	x, y := 0, 0

	for i, r := range testText {
		style := defStyle
		if i < len(typedText) {
			if rune(testText[i]) == typedText[i] {
				style = correctStyle
			} else {
				style = incorrectStyle
			}
		}

		if i == cursorPos {
			style = currentStyle
		}

		s.SetContent(x, y, r, nil, style)
		x += utf8.RuneLen(r)
		if x >= w {
			x = 0
			y++
		}
	}
}

func drawStats(s tcell.Screen, startTime time.Time, testStarted bool, typedText []rune, testText string, style tcell.Style, cursorPos int, errors int) (float64, float64) {
	_, h := s.Size()
	y := h - 2

	duration := time.Since(startTime)
	wpm := 0.0
	accuracy := 0.0

	if testStarted && duration.Seconds() > 0 {
		wordsTyped := float64(len(typedText)) / 5.0 // Assuming average word length of 5 characters
		wpm = (wordsTyped / duration.Seconds()) * 60.0

		// Calculate accuracy based on cumulative errors
		totalAttemptedChars := float64(cursorPos) // Use cursorPos as it represents how far the user has progressed in the test text

		if totalAttemptedChars > 0 {
			// Accuracy = (Total characters attempted - Cumulative errors) / Total characters attempted
			accuracy = (totalAttemptedChars - float64(errors)) / totalAttemptedChars * 100.0
			if accuracy < 0 { // Accuracy cannot be negative
				accuracy = 0
			}
		}
	}

	statLine := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%% | Errors: %d", wpm, accuracy, errors)
	printString(s, 0, y, statLine, style)

	// Progress bar
	progressBarWidth := 50
	progress := float64(cursorPos) / float64(len(testText))
	filledWidth := int(progress * float64(progressBarWidth))
	
	progressBar := "["
	for i := 0; i < filledWidth; i++ {
		progressBar += "#"
	}
	for i := filledWidth; i < progressBarWidth; i++ {
		progressBar += "-"
	}
	progressBar += "]"
	printString(s, 0, y+1, progressBar, style)

	return wpm, accuracy
}

func printString(s tcell.Screen, x, y int, str string, style tcell.Style) {
	for _, r := range str {
		s.SetContent(x, y, r, nil, style)
		x += utf8.RuneLen(r)
	}
}
