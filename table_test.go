package main

import "testing"

func Test_stringToSize(t *testing.T) {
	tests := []struct {
		name    string
		want    uint64
		wantErr bool
	}{
		{"0", 0, false},
		{"42", 42, false},

		{"1K", 1 << 10, false},
		{"2M", 2 << 20, false},
		{"3G", 3 << 30, false},
		{"4T", 4 << 40, false},
		{"5P", 5 << 50, false},
		{"6E", 6 << 60, false},

		{"", 0, true},
		{"abc", 0, true},
		{"10Z", 0, true},
		{"-5", 0, true},
		{"18446744073709551615", ^uint64(0), false},

		{" 10K", 0, true},
		{"10K ", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringToSize(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("stringToSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stringToSize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkStringToSize(b *testing.B) {
	cases := []string{
		"42",
		"1K",
		"128M",
		"512G",
		"2T",
		"5P",
		"6E",
		"invalid",
	}

	for _, input := range cases {
		b.Run(input, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = stringToSize(input)
			}
		})
	}
}
