package utils

import (
	"testing"
)

func BenchmarkRandString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandString(6)
	}
}

func TestTrimAllWhitespace(t *testing.T) {
	jsonStr :=
		`
	{
		"userId": "6286425165815353344",
		"category": 1
	}
	`
	t.Logf("TestTrimAllWhitespace jsonStr: %v", jsonStr)
	t.Logf("TestTrimAllWhitespace trimStr: %v", TrimAllWhitespace(jsonStr))
}
