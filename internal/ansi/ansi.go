package ansi

import "fmt"

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

func ResetScreen() {
	fmt.Print("\033[2J") // Clear screen
	fmt.Print("\033[H")  // Move cursor to top-left
}

func WriteChar(row, col int, char byte) {
	fmt.Printf("\033[%d;%dH%s", row, col, string(char))
}

func Backspace() {
	fmt.Print("\b \b")
}
