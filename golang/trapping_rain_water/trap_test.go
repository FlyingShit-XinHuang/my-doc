package trap

import (
	"testing"
)

func TestTrap(t *testing.T) {
	cases := []struct {
		input    []int
		expected int
	}{
		{
			[]int{1, 2, 3, 4, 5},
			0,
		},
		{
			[]int{5, 4, 3, 2, 1},
			0,
		},
		{
			[]int{5, 4, 3, 5},
			3,
		},
		{
			[]int{0, 1, 0, 2, 1, 0, 1, 3, 2, 1, 2, 1},
			6,
		},
	}

	for i, c := range cases {
		got := Trap(c.input)
		if got != c.expected {
			t.Fatalf("Case %d fail, expected: %d, got: %d", i, c.expected, got)
		}
	}
}
