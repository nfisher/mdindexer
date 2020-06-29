package edit

import (
	"testing"
)

var Dist = 0

func BenchmarkEditDistance(b *testing.B) {
	var dist = 0
	for i := 0; i < b.N; i++ {
		dist = Distance("intention", "execution")
	}
	Dist = dist
}

func BenchmarkEditDistance2(b *testing.B) {
	var dist = 0
	for i := 0; i < b.N; i++ {
		dist = Distance2("intention", "execution")
	}
	Dist = dist
}

var testCases = map[string]struct {
	s1   string
	s2   string
	dist int
}{
	"case1":  {"h", "hello", 4},
	"case1b": {"hello", "h", 4},
	"case2":  {"intention", "execution", 8},
	"case3":  {"kitten", "sitting", 5},
	"case4":  {"Saturday", "Sunday", 4},
	"case5":  {"hello", "hello", 0},
	"case6":  {"hello", "hell", 1},
}

func Test_min_edit_distance(t *testing.T) {
	for n, tc := range testCases {
		tc := tc
		t.Run(n, func(t *testing.T) {
			dist := Distance(tc.s1, tc.s2)
			if dist != tc.dist {
				t.Errorf("Distance(%s, %s)=%d, want %d", tc.s1, tc.s2, dist, tc.dist)
			}
		})
	}
}

func Test_min_edit_distance_v2(t *testing.T) {
	for n, tc := range testCases {
		tc := tc
		t.Run(n, func(t *testing.T) {
			dist := Distance2(tc.s1, tc.s2)
			if dist != tc.dist {
				t.Errorf("Distance(%s, %s)=%d, want %d", tc.s1, tc.s2, dist, tc.dist)
			}
		})
	}
}
