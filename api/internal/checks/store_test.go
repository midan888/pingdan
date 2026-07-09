package checks

import "testing"

func TestPctIdx(t *testing.T) {
	cases := []struct {
		n, pct int
		want   int
	}{
		{1, 50, 0},
		{4, 50, 2},
		{4, 95, 3},
		{10, 100, 9},
	}
	for _, c := range cases {
		if got := pctIdx(c.n, c.pct); got != c.want {
			t.Errorf("pctIdx(%d, %d) = %d, want %d", c.n, c.pct, got, c.want)
		}
	}
}

func TestSortInts(t *testing.T) {
	got := []int{50, 10, 30, 30, 20}
	sortInts(got)

	want := []int{10, 20, 30, 30, 50}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sortInts() = %#v, want %#v", got, want)
		}
	}
}
