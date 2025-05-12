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
	NEWLINE   byte = 10
)

type position struct {
	row, col int
}

type cursor struct {
	currPos, currWordPos, minPos position
}

type tracker struct {
	validWordsCount, validKeysCount, totalKeyStrokes int
	validKeysStreaks                                 []int
	// correctionCount                                                    int
}

type Test struct {
	expected     []string
	input        [][]byte
	cursor       cursor
	tracker      tracker
	durationMins float32
}

func (pos *position) isLessThan(pos2 position) bool {
	if pos.row < pos2.row {
		return true
	} else if pos.row == pos2.row && pos.col < pos2.col {
		return true
	} else {
		return false
	}
}

func (cursor *cursor) writeCharWithColor(inputChar byte, colorEscapeCode string) {
	ansi.WriteCharWithColor(cursor.currPos.row, cursor.currPos.col, inputChar, colorEscapeCode)
}

func (tracker *tracker) recordValidKey() {
	tracker.validKeysCount++
	tracker.validKeysStreaks[len(tracker.validKeysStreaks)-1]++
}

func (tracker *tracker) recordInvalidKey() {
	tracker.validKeysStreaks = append(tracker.validKeysStreaks, 0)
}

func NewTest(expected_str string) Test {
	linesArr := strings.Split(expected_str, "\n")

	test := Test{
		expected: linesArr,
		input:    make([][]byte, len(linesArr)),
		cursor:   cursor{},
		tracker:  tracker{validKeysStreaks: []int{0}},
	}

	for i := range test.input {
		test.input[i] = make([]byte, len(test.expected[i]))
		test.expected[i] += ""
	}
	return test
}

func (test *Test) getExpectedChar() byte {
	return test.expected[test.cursor.currPos.row][test.cursor.currPos.col]
}

func (test *Test) handleSpace() {
	expectedChar := test.getExpectedChar()

	if expectedChar != SPACE {
		test.cursor.writeCharWithColor(expectedChar, ansi.Red)
		test.tracker.recordInvalidKey()
		return
	}

	//pre process updatwa in case this last space
	tempRow := test.cursor.currPos.row
	tempCol := test.cursor.currPos.col + 1
	if test.cursor.currPos.col+1 == len(test.expected[test.cursor.currPos.row]) {
		tempRow++
		tempCol = 1 //might be 1
	}

	test.tracker.recordValidKey()

	if test.cursor.currWordPos.isLessThan(test.cursor.currPos) {
		if string(test.input[test.cursor.currPos.row][test.cursor.currWordPos.col:test.cursor.currPos.col]) == test.expected[test.cursor.currPos.row][test.cursor.currWordPos.col:test.cursor.currPos.col] {
			test.cursor.minPos = test.cursor.currPos
			test.tracker.validWordsCount++
		}
		test.cursor.currPos = position{row: tempRow, col: tempCol}
		test.cursor.currWordPos = test.cursor.currPos
	} else {
		test.cursor.currPos.col = tempCol
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

	isEnd := test.cursor.currPos.row >= len(test.expected) && test.cursor.currPos.col >= len(test.expected[test.cursor.currPos.row])
	isAllowedAtEnd := input == BACKSPACE || input == CTRLC

	if isEnd && !isAllowedAtEnd {
		return nil
	}

	// wrong char on expected SPACE
	if !isAllowedAtEnd {
		if expected := test.getExpectedChar(); expected == SPACE && input != expected {
			test.input[test.cursor.currPos.row][test.cursor.currPos.col] = byte(input)
			test.cursor.currPos.col++
			test.cursor.writeCharWithColor(input, ansi.Red)
			test.tracker.recordInvalidKey()
			return nil
		}
	}

	switch input {
	case CTRLC: // handle end test
		// if test.cursor.pos == len(test.expected)-1 && string(test.input[test.cursor.currWordPos:test.cursor.pos]) == test.expected[test.cursor.currWordPos:test.cursor.pos] {
		// 	test.tracker.validWordsCount++
		// }
		return errors.New("closing test")

	case SPACE:
		if !isEnd {
			test.tracker.totalKeyStrokes++
			test.input[test.cursor.currPos.row][test.cursor.currPos.col] = byte(SPACE)
			test.handleSpace()
		}

	case BACKSPACE: // handle backspace
		if test.cursor.minPos.isLessThan(test.cursor.currPos) {
			test.tracker.totalKeyStrokes++
			test.cursor.currPos.col--
			ansi.BackspaceAndReplace(test.getExpectedChar())
		}

	default: // handle normal input
		test.tracker.totalKeyStrokes++
		test.input[test.cursor.currPos.row][test.cursor.currPos.col] = byte(SPACE)
		test.cursor.currPos.col++
		expected := test.expected[test.cursor.currPos.row][test.cursor.currPos.col-1]
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
	fmt.Println(strings.Join(test.expected, "\n"))
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
	errors = len(test.tracker.validKeysStreaks) - 1
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
	fmt.Printf("consistency: %d%%\n", round(consistency*100))
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
