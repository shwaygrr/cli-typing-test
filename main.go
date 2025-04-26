package main

import "github.com/shwaygrr/cli-typing-test/internal/test"

func main() {
	// test_string := "could such"
	test_string := "could such old out begin great head world other give change thing play little any however problem now work"
	tt := test.NewTest(test_string)

	tt.RunTest()
}
