package test

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	CTRLC     byte = 3
	BACKSPACE byte = 127
	ENTER     byte = 13
	// LEFT_ARROW = 13
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
	switch input {
	case CTRLC, ENTER: //test prgogram enders
		os.Exit(1)
	case BACKSPACE: //handle backspace
		if test.cursorPos > 0 {
			test.input[test.cursorPos] = ' '                                        // Erase character
			fmt.Printf("\033[1;%dH \033[1;%dH", test.cursorPos+1, test.cursorPos+1) // Move cursor back,

			test.cursorPos--
		}
	default: //normal input
		if test.cursorPos < len(test.expected)-1 {
			test.input[test.cursorPos] = byte(input)
			fmt.Printf("\033[1;%dH%s", test.cursorPos+1, string(input))
			test.cursorPos++
		}
	}
}

func (test *Test) termSetup() {
	fmt.Print("\033[2J") // Clear screen
	fmt.Print("\033[H")  // Move cursor to top-left
	fmt.Println(test.expected)
	fmt.Printf("\033[1;1H") // Move to row 1, column 1
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
