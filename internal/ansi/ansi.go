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

func WriteCharWithColor(row, col int, char byte, colorEscapeCode string) {
	fmt.Printf("%s\033[%d;%dH%s", colorEscapeCode, row, col, string(char))
}

func ChangeTextColor(color string) {
	fmt.Print(color)
}

func Backspace() {
	fmt.Print("\b \b")
}
