package main

import (
	"fmt"
	"testing"
)

func TestLamport(t *testing.T) {

	fmt.Println(lc.Time())

	fmt.Println(lc.Increment())

	lc.Witness(LamportTime(lc.Time()))

	fmt.Println(lc.Time())

	lc.Witness(LamportTime(20))

	fmt.Println(lc.Time())

	// lc.Increment()
	lc.Witness(LamportTime(5))

	fmt.Println(lc.Time())

	lc.Witness(LamportTime(22))

	fmt.Println(lc.Time())
}
