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
)

type Test struct {
	expected                 string
	input                    []byte
	cursorPos                int
	total_chars, total_words int
	// wpm, cpm, accuracy float32
}

func NewTest(expected_str string) Test {
	test := Test{
		expected:    expected_str,
		input:       make([]byte, len(expected_str)),
		cursorPos:   0,
		total_chars: len(expected_str),
		total_words: len(strings.Trim(expected_str, " ")),

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

func (test *Test) handleInput(input byte) error {
	isAllowedInput := ('A' <= input && input <= 'Z') ||
		('a' <= input && input <= 'z') ||
		('0' <= input && input <= '9') ||
		strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", rune(input)) ||
		input == ' ' ||
		input == CTRLC ||
		input == BACKSPACE

	if !isAllowedInput {
		return nil // not error but does nothing with input
	}

	switch input {
	case CTRLC: // handle end test
		return errors.New("closing test")

	case BACKSPACE: // handle backspace
		if test.cursorPos > 0 {
			test.cursorPos--
			ansi.BackspaceAndReplace(test.getExpectedChar())
		}

	default: // handle normal input
		if test.cursorPos < len(test.expected) {
			test.input[test.cursorPos] = byte(input)
			test.cursorPos++
			if input == test.expected[test.cursorPos-1] {
				ansi.WriteCharWithColor(1, test.cursorPos, input, ansi.Green)
			} else {
				ansi.WriteCharWithColor(1, test.cursorPos, input, ansi.Red)
			}
		}
	}
	return nil
}

func (test *Test) termSetup() {
	ansi.ResetScreen()
	ansi.ChangeTextColor(ansi.Cyan)
	fmt.Println(test.expected)
	ansi.WriteCharWithColor(1, 1, 0, "") //move to start and
}

func (test *Test) timeCalcs(duration_minutes float32) (float32, float32) {
	valid_chars := test.total_chars - test.calcInputCharDiff()
	cpm := float32(valid_chars) / duration_minutes
	wpm := float32(test.total_words) / duration_minutes

	fmt.Println(ansi.Reset+"cpm:", cpm)
	fmt.Println("\nduration:", duration_minutes)
	fmt.Println("wpm:", wpm)
	fmt.Println("valid_chars:", valid_chars)
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
