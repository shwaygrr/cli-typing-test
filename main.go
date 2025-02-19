package main

import "github.com/shwaygrr/cli-typing-test/internal/test"

func main() {
	test_string := "In Go, there are several ways to determine the type of a value."
	tt := test.NewTest(test_string)

	tt.RunTest()
}
