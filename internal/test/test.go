package test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"math"

	"github.com/shwaygrr/cli-typing-test/internal/ansi"
	"golang.org/x/term"
)

const (
	CTRLC     byte = 3
	BACKSPACE byte = 127
	ENTER     byte = 13
	SPACE     byte = 32
)

type cursor struct {
	pos, currWordPos, minPos int
}

type tracker struct {
	validWordsCount, validKeysCount, invalidKeysCount, totalKeyStrokes int
	validKeysStreaks                                                   []int
	// correctionCount                                                    int
}

type Test struct {
	expected     string
	input        []byte
	cursor       cursor
	tracker      tracker
	durationMins float32
}

func (cursor *cursor) writeCharWithColor(inputChar byte, colorEscapeCode string) {
	ansi.WriteCharWithColor(1, cursor.pos, inputChar, colorEscapeCode)
}

func (tracker *tracker) recordValidKey() {
	streaksLastIndex := len(tracker.validKeysStreaks) - 1 //should be the same as amount of mistakes (needs refactor)
	tracker.validKeysCount++
	tracker.validKeysStreaks[streaksLastIndex]++
}

func (tracker *tracker) recordInvalidKey() {
	tracker.invalidKeysCount++
	tracker.validKeysStreaks = append(tracker.validKeysStreaks, 0)
}

func NewTest(expected_str string) Test {
	test := Test{
		expected: expected_str,
		input:    make([]byte, len(expected_str)),
		cursor:   cursor{},
		tracker:  tracker{validKeysStreaks: []int{0}},
	}
	return test
}

func (test *Test) getExpectedChar() byte {
	return test.expected[test.cursor.pos]
}

func (test *Test) handleSpace() {
	if test.expected[test.cursor.pos] != SPACE {
		test.cursor.writeCharWithColor(test.expected[test.cursor.pos], ansi.Red)
		test.tracker.recordInvalidKey()
		return
	}

	test.tracker.recordValidKey()

	if test.cursor.currWordPos < test.cursor.pos {
		if string(test.input[test.cursor.currWordPos:test.cursor.pos]) == test.expected[test.cursor.currWordPos:test.cursor.pos] {
			test.cursor.minPos = test.cursor.pos
			test.tracker.validWordsCount++
		}
		test.cursor.pos++
		test.cursor.currWordPos = test.cursor.pos
	} else {
		test.cursor.pos++
	}

	test.cursor.writeCharWithColor(SPACE, ansi.Green)
}

func (test *Test) handleInput(input byte) error {
	isAllowedInput := ('A' <= input && input <= 'Z') ||
		('a' <= input && input <= 'z') ||
		('0' <= input && input <= '9') ||
		strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", rune(input)) ||
		// input == ' ' ||
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
			test.cursor.writeCharWithColor(input, ansi.Red)
			test.tracker.recordInvalidKey()
			return nil
		}
	}

	switch input {
	case CTRLC: // handle end test
		if string(test.input[test.cursor.currWordPos:test.cursor.pos]) == test.expected[test.cursor.currWordPos:test.cursor.pos] {
			test.tracker.validWordsCount++
		}
		return errors.New("closing test")

	case SPACE:
		if test.cursor.pos < len(test.expected) {
			test.tracker.totalKeyStrokes++
			test.input[test.cursor.pos] = byte(SPACE)
			test.handleSpace()
		}

	case BACKSPACE: // handle backspace
		if test.cursor.pos > test.cursor.minPos {
			test.tracker.totalKeyStrokes++
			test.cursor.pos--
			ansi.BackspaceAndReplace(test.getExpectedChar())
		}

	default: // handle normal input
		test.tracker.totalKeyStrokes++
		test.input[test.cursor.pos] = byte(input)
		test.cursor.pos++
		expected := test.expected[test.cursor.pos-1]
		if input == expected {
			test.cursor.writeCharWithColor(expected, ansi.Green)
			test.tracker.recordValidKey()
		} else {
			test.cursor.writeCharWithColor(expected, ansi.Red)
			test.tracker.recordInvalidKey()
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

// func (test *Test) calcInputDiffs() int {
// 	diff_count := 0

// 	for i := range test.expected {
// 		if test.expected[i] != test.input[i] {
// 			diff_count++
// 		}
// 	}

// 	return diff_count
// }

func (test *Test) computeMetrics() (accuracy, cpm, wpm, consistency float32, errors int) {
	max := func(arr []int) int {
		max := 0
		for _, num := range arr {
			if num > max {
				max = num
			}
		}
		return max
	}

	accuracy = float32(test.tracker.validKeysCount) / float32(test.tracker.totalKeyStrokes)
	cpm = float32(test.tracker.validKeysCount) / test.durationMins
	wpm = float32(test.tracker.validWordsCount) / test.durationMins

	consistency = float32(max(test.tracker.validKeysStreaks)) / float32(test.tracker.totalKeyStrokes)
	errors = test.tracker.invalidKeysCount
	// corrections = ...
	return
}

func (test *Test) endTest(startTime time.Time) {
	round := func(float float32) int {
		return int(math.Round(float64(float)))
	}
	//handle tottal key strokes = 0
	//hgandle time 0

	test.durationMins = float32(time.Since(startTime).Minutes())
	accuracy, cpm, wpm, consistency, errors := test.computeMetrics()

	duration_type := "m"
	duration := test.durationMins

	if duration < 1 {
		duration *= 60
		duration_type = "s"
	}

	fmt.Printf("%s\ntime:%d%s\n", ansi.Reset, round(duration), duration_type)
	fmt.Println("cpm:", round(cpm))
	fmt.Println("wpm:", round(wpm))
	fmt.Printf("accuracy: %d%%\n", round(accuracy*100))
	fmt.Printf("cosistency: %d%%\n", round(consistency*100))
	fmt.Println("errors:", errors)

	fmt.Printf("\n%+v\n", test.tracker)
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
	test.endTest(startTime)
}
