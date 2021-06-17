package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestSecondTimeStringToNanoTime(t *testing.T) {
	times, err := SecondTimeStringToNanoTime("2021-03-25 17:56:27")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(times/int64(time.Microsecond))
}
