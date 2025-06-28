package main

import (
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

func runTest(s tcell.Screen, textToType string, defStyle, correctStyle, incorrectStyle, currentStyle tcell.Style, username string) (float64, float64) {
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
					SaveScore(username, wpm, accuracy)
				}
				return wpm, accuracy
			}
		}
	}
	return 0, 0 // Should not be reached, but required for function signature
}
