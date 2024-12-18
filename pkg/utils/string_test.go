package utils

import "testing"

// TestIsNilOrEmpty The function is a unit test function for the IsNilOrEmpty function.
//
// It contains multiple test cases to test whether the behavior of the IsNilOrEmpty function meets expectations under various conditions.
func TestIsNilOrEmpty(t *testing.T) {
	type args struct {
		value *string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{
				value: nil,
			},
			want: true,
		},
		{
			name: "Empty string",
			args: args{
				value: func() *string { s := ""; return &s }(),
			},
			want: true,
		},
		{
			name: "A string containing only spaces",
			args: args{
				value: func() *string { s := "   "; return &s }(),
			},
			want: true,
		},
		{
			name: "Non-empty string",
			args: args{
				value: func() *string { s := "hello"; return &s }(),
			},
			want: false,
		},
		{
			name: "Non-empty string containing spaces",
			args: args{
				value: func() *string { s := " hello world "; return &s }(),
			},
			want: false,
		},
	}
	// Traverse all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the tested function and compare whether the actual result is consistent with the expected result.
			if got := IsNilOrEmpty(tt.args.value); got != tt.want {
				t.Errorf("IsNilOrEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
