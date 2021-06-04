package main

import (
	"errors"
	"fmt"

	"github.com/elprans/go-promise"
)

func main() {
	var p1 = promise.Reject(errors.New("bad error"))
	_, err := p1.Await()
	fmt.Println(err)
}
