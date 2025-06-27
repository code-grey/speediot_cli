package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

const (
	textToType = "The quick brown fox jumps over the lazy dog. This is a typing test application written in Go. Practice makes perfect. Improve your typing speed and accuracy. Have fun!"
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

	runTest(s, defStyle, correctStyle, incorrectStyle, currentStyle)

	s.Fini()
}

func runTest(s tcell.Screen, defStyle, correctStyle, incorrectStyle, currentStyle tcell.Style) {
	rand.Seed(time.Now().UnixNano())
	words := strings.Fields(textToType)
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
	testText := strings.Join(words, " ")

	typedText := make([]rune, 0, len(testText))
	cursorPos := 0
	errors := 0
	startTime := time.Time{}
	testStarted := false

	for {
		s.Clear()
		printText(s, testText, typedText, cursorPos, correctStyle, incorrectStyle, currentStyle, defStyle)
		drawStats(s, startTime, testStarted, typedText, testText, errors, defStyle)
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
						errors++
					}
				}
				cursorPos += utf8.RuneLen(ev.Rune())
			}

			if cursorPos >= len(testText) && testStarted {
				// Test finished
				s.Clear()
				printText(s, testText, typedText, cursorPos, correctStyle, incorrectStyle, currentStyle, defStyle)
				drawStats(s, startTime, testStarted, typedText, testText, errors, defStyle)
				s.Show()
				time.Sleep(5 * time.Second) // Display results for a few seconds
				return
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

func drawStats(s tcell.Screen, startTime time.Time, testStarted bool, typedText []rune, testText string, errors int, style tcell.Style) {
	w, h := s.Size()
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

	statLine := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%% | Errors: %d", wpm, accuracy, errors)
	printString(s, 0, y, statLine, style)
}

func printString(s tcell.Screen, x, y int, str string, style tcell.Style) {
	for _, r := range str {
		s.SetContent(x, y, r, nil, style)
		x += utf8.RuneLen(r)
	}
}
