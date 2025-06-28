package main

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

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
