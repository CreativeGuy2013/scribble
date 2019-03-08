package mkdir

import (
	"math/rand"
	"os"
	"testing"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateData(n int) (data []string) {
	data = make([]string, n)

	for i := 0; i < n; i++ {
		data[i] = randSeq(10)
	}

	return
}
func BenchmarkMkdirAll(b *testing.B) {
	data := generateData(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		os.MkdirAll("./db/db/db/db/db/db/"+data[i], 0755)
	}

	b.StopTimer()
	os.RemoveAll("./db")
}

