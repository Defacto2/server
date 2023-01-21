package html3

import "testing"

func TestLimit(t *testing.T) {
	type args struct {
		count int
		limit int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"zero", args{0, 0}, 0},
		{"1-1", args{1, 1}, 1},
		{"2-1", args{2, 1}, 1},
		{"1-2", args{1, 2}, 2},
		{"1 page", args{99, 100}, 100},
		{"2 pages", args{151, 100}, 100},
		{"5 pages", args{501, 100}, 100},
		{"1 page extended", args{101, 100}, 150},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Limit(tt.args.count, tt.args.limit); got != tt.want {
				t.Errorf("Limit() = %v, want %v", got, tt.want)
			}
		})
	}
}
