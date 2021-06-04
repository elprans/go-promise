package main

import (
	"fmt"

	"github.com/elprans/go-promise"
)

func main() {
	var p = promise.Resolve(nil).
		Then(func(data promise.Any) (promise.Any, error) {
			fmt.Println("I will execute first")
			return nil, nil
		}).
		Then(func(data promise.Any) (promise.Any, error) {
			fmt.Println("And I will execute second!")
			return nil, nil
		}).
		Then(func(data promise.Any) (promise.Any, error) {
			fmt.Println("Oh I'm last :(")
			return nil, nil
		})

	p.Await()
}
