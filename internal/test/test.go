package test

import (
	"fmt"
	"os"

	"github.com/shwaygrr/cli-typing-test/internal/ansi"
	"golang.org/x/term"
)

const (
	CTRLC      byte = 3
	BACKSPACE  byte = 127
	ENTER      byte = 13
	LEFT_ARROW      = 13
	// RIGHT_ARROW =
)

type Test struct {
	expected  string
	input     []byte
	cursorPos int
}

func NewTest(expected_str string) Test {
	test := Test{
		expected:  expected_str,
		input:     make([]byte, len(expected_str)),
		cursorPos: 0,
	}
	return test
}

func (test *Test) handleInput(input byte) {
	isAllowedInput := ('A' <= input && input <= 'Z') || ('a' <= input && input <= 'z') || ('0' <= input && input <= '9') || input == ' ' || input == CTRLC || input == BACKSPACE

	if !isAllowedInput {
		return
	}

	switch input {
	case CTRLC: // handle end test
		os.Exit(1)
	case BACKSPACE: //handle backspace
		if test.cursorPos > 0 {
			test.cursorPos--
			ansi.Backspace()
		}
	default: //handle normal input
		if test.cursorPos < len(test.expected) {
			test.input[test.cursorPos] = byte(input)
			test.cursorPos++
			ansi.WriteChar(1, test.cursorPos, input)
		}
	}
}

func (test *Test) termSetup() {
	ansi.ResetScreen()
	fmt.Println(test.expected)
	ansi.WriteChar(1, 1, 0)
}

func (test *Test) RunTest() {
	var file_descriptor = int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(file_descriptor)
	if err != nil {
		panic(err)
	}
	defer term.Restore(file_descriptor, oldState)

	test.termSetup()

	for {
		// fmt.Print(string(test.expected[test.cursorPos]))
		b := make([]byte, 1)
		_, err = os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}

		test.handleInput(b[0])
	}
}
