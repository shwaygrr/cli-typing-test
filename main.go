package main

import "github.com/shwaygrr/cli-typing-test/internal/test"

func main() {
	test_string := "hello world"
	tt := test.NewTest(test_string)

	tt.RunTest()
}
