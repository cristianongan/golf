package utils

// unit test for utils package
// func name must be Testxxx (xxx is )
import (
	"testing"
)

func TestCount(t *testing.T) {
	tables := []struct {
		a []int
		n int
	}{
		{[]int{1, 1}, 2},
		{[]int{1, 2}, 3},
		{[]int{2, 2}, 4},
		{[]int{5, 2}, 7},
	}

	for _, table := range tables {
		total := Sum(table.a)
		if total != table.n {
			t.Errorf("Sum of (%d+%d) was incorrect, got: %d, want: %d.", table.a[0], table.a[1], total, table.n)
		}
	}
}
