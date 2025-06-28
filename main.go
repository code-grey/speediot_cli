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
		runTest(s, selectedText, defStyle, correctStyle, incorrectStyle, currentStyle)
		if !askToRestart(s, defStyle) {
			break
		}
	}

	s.Fini()
}

func runTest(s tcell.Screen, textToType string, defStyle, correctStyle, incorrectStyle, currentStyle tcell.Style) {
	rand.Seed(time.Now().UnixNano())
	words := strings.Fields(textToType)
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
	testText := strings.Join(words, " ")

	typedText := make([]rune, 0, len(testText))
	cursorPos := 0
	startTime := time.Time{}
	testStarted := false

	for {
		s.Clear()
		printText(s, testText, typedText, cursorPos, correctStyle, incorrectStyle, currentStyle, defStyle)
		drawStats(s, startTime, testStarted, typedText, testText, defStyle, cursorPos)
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
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
						// No longer incrementing errors here, calculated dynamically in drawStats
					}
				}
				cursorPos += utf8.RuneLen(ev.Rune())
			}

			if cursorPos >= len(testText) && testStarted {
				// Test finished
				s.Clear()
				printText(s, testText, typedText, cursorPos, correctStyle, incorrectStyle, currentStyle, defStyle)
				drawStats(s, startTime, testStarted, typedText, testText, defStyle, cursorPos)
				s.Show()
				return
			}
		}
	}
}

func showMainMenu(s tcell.Screen, style tcell.Style) string {
	asciiArt := []string{
		`  _______ _            _   _             _   _           _    `,
		` |__   __| |          | | (_)           | | (_)         | |   `,
		`    | |  | |__   ___  | |_ _ _ __ ___   | |_ _  ___  ___| | __`,
		`    | |  | '_ \ / _ \ | __| | '_ ` + "`" + ` _ \  | __| |/ _ \/ __| |/ /`,
		`    | |  | | | |  __/ | |_| | | | | | | | |_| |  __/ (__|   < `,
		`    |_|  |_| |_|\___|  \__|_|_|_| |_| |_|  \__|_|\___|\___|_|\_\`,
		`                                                             `,
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

func askToRestart(s tcell.Screen, style tcell.Style) bool {
	for {
		s.Clear()
		printString(s, 0, 0, "Play again? (y/n)", style)
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyRune {
				if ev.Rune() == 'y' || ev.Rune() == 'Y' {
					return true
				} else if ev.Rune() == 'n' || ev.Rune() == 'N' {
					return false
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

func drawStats(s tcell.Screen, startTime time.Time, testStarted bool, typedText []rune, testText string, style tcell.Style, cursorPos int) {
	_, h := s.Size()
	y := h - 2

	duration := time.Since(startTime)
	wpm := 0.0
	accuracy := 0.0

	if testStarted && duration.Seconds() > 0 {
		wordsTyped := float64(len(typedText)) / 5.0 // Assuming average word length of 5 characters
		wpm = (wordsTyped / duration.Seconds()) * 60.0

		correctChars := 0
		for i, r := range typedText {
			if i < len(testText) && r == rune(testText[i]) {
				correctChars++
			}
		}
		if len(typedText) > 0 {
			accuracy = (float64(correctChars) / float64(len(typedText))) * 100.0
		}
	}

	incorrectChars := 0
	for i, r := range typedText {
		if i < len(testText) && r != rune(testText[i]) {
			incorrectChars++
		}
	}
	statLine := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%% | Errors: %d", wpm, accuracy, incorrectChars)
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
}

func printString(s tcell.Screen, x, y int, str string, style tcell.Style) {
	for _, r := range str {
		s.SetContent(x, y, r, nil, style)
		x += utf8.RuneLen(r)
	}
}
