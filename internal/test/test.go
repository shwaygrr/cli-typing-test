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

type Cursor struct {
	pos, currWordPos, minPos int
}

type Test struct {
	expected               string
	input                  []byte
	cursor                 Cursor
	totalChars, totalWords int
	// wpm, cpm, accuracy float32
}

func NewTest(expected_str string) Test {
	test := Test{
		expected: expected_str,
		input:    make([]byte, len(expected_str)),
		cursor: Cursor{
			pos:         0,
			currWordPos: 0,
			minPos:      0,
		},
		totalChars: len(expected_str),
		totalWords: len(strings.Trim(expected_str, " ")),
		// wpm:       0,
		// cpm:       0,
		// accuracy:  0,
	}
	return test
}

func (test *Test) getExpectedChar() byte {
	return test.expected[test.cursor.pos]
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

	if test.expected[test.cursor.pos] != SPACE {
		ansi.WriteCharWithColor(1, test.cursor.pos, test.expected[test.cursor.pos], ansi.Red)
		return
	}

	if test.cursor.currWordPos < test.cursor.pos {
		if string(test.input[test.cursor.currWordPos:test.cursor.pos]) == test.expected[test.cursor.currWordPos:test.cursor.pos] {
			test.cursor.minPos = test.cursor.pos
		}
		test.cursor.pos++
		test.cursor.currWordPos = test.cursor.pos
	} else {
		test.cursor.pos++
	}

	ansi.WriteCharWithColor(1, test.cursor.pos, SPACE, ansi.Green)
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

	if test.cursor.pos >= len(test.expected) && !isAllowedAtEnd {
		return nil
	}

	// wrong char on expected SPACE
	if !isAllowedAtEnd {
		if expected := test.getExpectedChar(); expected == SPACE && input != expected {
			test.input[test.cursor.pos] = byte(input)
			test.cursor.pos++
			ansi.WriteCharWithColor(1, test.cursor.pos, input, ansi.Red)
			return nil
		}
	}

	switch input {
	case CTRLC: // handle end test
		return errors.New("closing test")

	case SPACE:
		if test.cursor.pos < len(test.expected) {
			test.input[test.cursor.pos] = byte(SPACE)
			test.handleSpace()
		}

	case BACKSPACE: // handle backspace
		if test.cursor.pos > test.cursor.minPos {
			test.cursor.pos--
			ansi.BackspaceAndReplace(test.getExpectedChar())
		}

	default: // handle normal input
		test.input[test.cursor.pos] = byte(input)
		test.cursor.pos++
		expected := test.expected[test.cursor.pos-1]
		if input == expected {
			ansi.WriteCharWithColor(1, test.cursor.pos, expected, ansi.Green)
		} else {
			ansi.WriteCharWithColor(1, test.cursor.pos, expected, ansi.Red)
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
