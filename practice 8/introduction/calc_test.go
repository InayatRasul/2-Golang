package calc
import "testing"

// func TestAdd(t *testing.T) {
// 	got := Add(2, 3)
// 	want := 6
// 	if got != want {
// 		t.Errorf("Add(2, 3) = %d; want %d", got, want)
// 	}
// }

func TestAddTableDriven(t *testing.T) {
	tests := []struct {
		// Define a struct for each test case and create a slice of them
		name string
		a, b int
		want int
	}{
		{"both positive", 2, 3, 5},
		{"positive + zero", 5, 0, 5},
		{"negative + positive", -1, 4, 3},
		{"both negative", -2, -3, -5},
	}
	for _, tt := range tests {
		// Loop over each test case
		t.Run(tt.name, func(t *testing.T) {
			// Run each case as a subtest
			got := Add(tt.a, tt.b)
			if got != tt.want {
				// Check the result
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got,tt.want) // Report failure if it doesn't match
			}
		})
	}
}