package server

import "testing"

func Test_VSCodeCors(t *testing.T) {
	type testCase struct {
		port     int
		origin   string
		expected bool
	}

	cases := []testCase{
		{
			port:     3000,
			origin:   "http://v--19cf5ppcsqee9rrkfifq1ajc8b7nv2t96593a6n6bn95st339ul8.localhost:3000",
			expected: true,
		},
		{
			port:     3000,
			origin:   "http://v--1f1nq97ha8fjnonnlusc36olp1p7do9ddp5bnr0apu6pt4phaoq0.localhost:3000",
			expected: true,
		},
		{
			port:     3000,
			origin:   "http://localhost:3000",
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.origin, func(t *testing.T) {
			vscodeCors, err := NewVscodeCors(tc.port)
			if err != nil {
				t.Fatalf("Failed to create vscodeCors: %v", err)
			}
			if vscodeCors.allowOrigin(tc.origin) != tc.expected {
				t.Fatalf("Expected %v but got %v", tc.expected, !tc.expected)
			}
		})
	}
}
