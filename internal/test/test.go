package test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shwaygrr/cli-typing-test/internal/ansi"
	"golang.org/x/term"
)

const (
	CTRLC     byte = 3
	BACKSPACE byte = 127
	ENTER     byte = 13
	SPACE     byte = 32
)

type Test struct {
	expected                             string
	input                                []byte
	cursorPos, currWordPos, minCursorPos int
	totalChars, totalWords               int
	// wpm, cpm, accuracy float32
}

func NewTest(expected_str string) Test {
	test := Test{
		expected:     expected_str,
		input:        make([]byte, len(expected_str)),
		cursorPos:    0,
		currWordPos:  0,
		minCursorPos: 0,
		totalChars:   len(expected_str),
		totalWords:   len(strings.Trim(expected_str, " ")),
		// wpm:       0,
		// cpm:       0,
		// accuracy:  0,
	}
	return test
}

func (test *Test) getExpectedChar() byte {
	return test.expected[test.cursorPos]
}

func (test *Test) calcInputCharDiff() int {
	diff_count := 0

	for i := range test.expected {
		if test.expected[i] != test.input[i] {
			diff_count++
		}
	}

	return diff_count
}

func (test *Test) handleSpace() {

	if test.expected[test.cursorPos] != SPACE {
		ansi.WriteCharWithColor(1, test.cursorPos, test.expected[test.cursorPos], ansi.Red)
		return
	}

	if test.currWordPos < test.cursorPos {
		if string(test.input[test.currWordPos:test.cursorPos]) == test.expected[test.currWordPos:test.cursorPos] {
			test.minCursorPos = test.cursorPos
		}
		test.cursorPos++
		test.currWordPos = test.cursorPos
	} else {
		test.cursorPos++
	}

	ansi.WriteCharWithColor(1, test.cursorPos, SPACE, ansi.Green)
}

func (test *Test) handleInput(input byte) error {
	isAllowedInput := ('A' <= input && input <= 'Z') ||
		('a' <= input && input <= 'z') ||
		('0' <= input && input <= '9') ||
		strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", rune(input)) ||
		input == ' ' ||
		input == CTRLC ||
		input == SPACE ||
		input == BACKSPACE

	if !isAllowedInput {
		return nil
	}

	isAllowedAtEnd := input == BACKSPACE || input == CTRLC

	if test.cursorPos >= len(test.expected) && !isAllowedAtEnd {
		return nil
	}

	// wrong char on expected SPACE
	if !isAllowedAtEnd {
		if expected := test.getExpectedChar(); expected == SPACE && input != expected {
			test.input[test.cursorPos] = byte(input)
			test.cursorPos++
			ansi.WriteCharWithColor(1, test.cursorPos, input, ansi.Red)
			return nil
		}
	}

	switch input {
	case CTRLC: // handle end test
		return errors.New("closing test")

	case SPACE:
		if test.cursorPos < len(test.expected) {
			test.input[test.cursorPos] = byte(SPACE)
			test.handleSpace()
		}

	case BACKSPACE: // handle backspace
		if test.cursorPos > test.minCursorPos {
			test.cursorPos--
			ansi.BackspaceAndReplace(test.getExpectedChar())
		}

	default: // handle normal input
		test.input[test.cursorPos] = byte(input)
		test.cursorPos++
		expected := test.expected[test.cursorPos-1]
		if input == expected {
			ansi.WriteCharWithColor(1, test.cursorPos, expected, ansi.Green)
		} else {
			ansi.WriteCharWithColor(1, test.cursorPos, expected, ansi.Red)
		}
	}
	return nil
}

func (test *Test) termSetup() {
	ansi.ResetScreen()
	ansi.ChangeTextColor(ansi.Cyan)
	fmt.Println(test.expected)
	ansi.WriteCharWithColor(1, 1, 0, "") // move to start
}

func (test *Test) timeCalcs(duration_minutes float32) (float32, float32) {
	validChars := test.totalChars - test.calcInputCharDiff()
	cpm := float32(validChars) / duration_minutes
	wpm := float32(test.totalWords) / duration_minutes

	duration_type := "minutes"
	duration := duration_minutes

	if duration_minutes < 1 {
		duration = duration_minutes * 60
		duration_type = "seconds"
	}

	fmt.Println(ansi.Reset+"\ncpm:", cpm)
	fmt.Println("duration:", duration, duration_type)
	fmt.Println("wpm:", wpm)
	fmt.Println("valid_chars:", validChars)
	return cpm, wpm
}

func (test *Test) RunTest() {
	startTime := time.Now()
	// ticker := time.NewTicker(1 * time.Second)
	// go func() {
	// 	for range ticker.C {
	// 		elapsed := time.Since(startTime)
	// 		fmt.Printf("\rTime elapsed: %.2f seconds", elapsed.Seconds())
	// 	}
	// }()

	var file_descriptor = int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(file_descriptor)
	if err != nil {
		panic(err)
	}
	defer term.Restore(file_descriptor, oldState)

	test.termSetup()
	defer ansi.ChangeTextColor(ansi.Reset)

	for {
		b := make([]byte, 1)
		_, err = os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}

		err := test.handleInput(b[0])
		if err != nil {
			break
		}
	}

	test.timeCalcs(float32(time.Since(startTime).Minutes()))
}

//to make each string distinguashble without splitting and reading by spaces
//store all the words in an  array of strings (words).
//Need to then modify the printing of the words (handle input) to consider the new DS
