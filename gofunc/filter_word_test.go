package gofunc

import (
	"testing"
)

func BenchmarkFilterWord(b *testing.B) {

	BuildDict("./maskingwords.txt")
	for i := 0; i < b.N; i++ {

		FilterWord("习近平是个江泽民", true)
	}
}
