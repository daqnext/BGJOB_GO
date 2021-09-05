package test

import (
	"fmt"
	"testing"
)

func TestBigNumber(t *testing.T) {
	var x uint64 = 18446744073709551615
	x = x + 2
	fmt.Println(x)
}
