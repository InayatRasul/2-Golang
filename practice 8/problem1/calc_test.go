package calc

import "testing"

type testCase struct {
	name string
	a    int
	b    int
	want int
}

func TestSubtract_TableDriven(t *testing.T) {
	tests := []testCase{
		{"both positive", 5, 3, 2},
		{"positive minus zero", 5, 0, 5},
		{"negative minus positive", -2, 3, -5},
		{"both negative", -5, -3, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name        string
		a           int
		b           int
		want        int
		expectError bool
	}{
		{"normal division", 10, 2, 5, false},
		{"division by one", 10, 1, 10, false},
		{"negative dividend", -10, 2, -5, false},
		{"negative divisor", 10, -2, -5, false},
		{"both negative", -10, -2, 5, false},
		{"division by zero", 10, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %d, want %d", got, tt.want)
				}
			}
		})
	}
}
