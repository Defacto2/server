package html3_test

import (
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	l := html3.Limit
	assert.Equal(t, l(0, 0), 0)
	assert.Equal(t, l(1, 1), 1)
	assert.Equal(t, l(2, 1), 1)
	assert.Equal(t, l(1, 2), 2)
	assert.Equal(t, l(99, 100), 100)
	assert.Equal(t, l(151, 100), 100)
	assert.Equal(t, l(501, 100), 100)
	assert.Equal(t, l(101, 100), 150)
}

func TestPagi(t *testing.T) {
	type args struct {
		page    int
		maxPage uint
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
		want2 int
	}{
		{"empty", args{}, 0, 0, 0},
		{"1 page", args{1, 1}, 0, 0, 0},
		{"2 pages", args{1, 2}, 0, 0, 0},
		{"3 pages", args{1, 3}, 2, 0, 0},
		{"4 pages", args{1, 4}, 2, 3, 0},
		{"start of many pages", args{2, 10}, 2, 3, 4},
		{"middle of many pages", args{5, 10}, 4, 5, 6},
		{"near end of many pages", args{9, 10}, 7, 8, 9},
		{"last of many pages", args{10, 10}, 7, 8, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := html3.Pagi(tt.args.page, tt.args.maxPage)
			assert.Equal(t, got, tt.want, "value a")
			assert.Equal(t, got1, tt.want1, "value b")
			assert.Equal(t, got2, tt.want2, "value c")
		})
	}
}

func TestNavi(t *testing.T) {
	limit := 10
	page := 2
	maxPage := uint(5)
	current := "current"
	qs := "query"

	expected := html3.Navigate{
		Current:  current,
		Limit:    limit,
		Page:     page,
		PagePrev: 1,
		PageNext: 3,
		PageMax:  5,
		QueryStr: qs,
	}

	result := html3.Navi(limit, page, maxPage, current, qs)

	if result != expected {
		t.Errorf("Navi(%d, %d, %d, %s, %s) = %v; want %v", limit, page, maxPage, current, qs, result, expected)
	}
}

func TestList(t *testing.T) {
	// TODO: Write test cases for the List function
}

func TestSumBytesBy(t *testing.T) {
	// TODO: Write test cases for the sumBytesBy function
}

func TestQs(t *testing.T) {
	// TODO: Write test cases for the qs function
}

func TestPrevious(t *testing.T) {
	// TODO: Write test cases for the previous function
}

func TestNext(t *testing.T) {
	// TODO: Write test cases for the next function
}
