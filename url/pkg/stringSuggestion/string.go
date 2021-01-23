package stringSuggestion

import (
	cryptoRand "crypto/rand"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

var mutantTypes = [3]string{
	"deletion",
	"substitute",
	"insertion",
}

var dna = []string{
	"A",
	"C",
	"G",
	"T",
}

func Suggest(str string, maxStringEditLength int, maxSuggestedStringLength int) string {
	rand.Seed(time.Now().UnixNano())
	var generatedStr string
	if len(str) < maxStringEditLength {
		maxStringEditLength = len(str)
	}
	k := rand.Intn(maxStringEditLength) + 1
	for i := 0; i < k; i++ {
		n := rand.Int() % len(mutantTypes)
		switch mutantTypes[n] {
		case "deletion":
			generatedStr = deletion(str, k)
		case "substitute":
			generatedStr = substitute(str, k)
		case "insertion":
			generatedStr = insertion(str, k)
		}
	}
	if len(generatedStr) < maxSuggestedStringLength {
		return url.QueryEscape(generatedStr + randString(maxSuggestedStringLength-len(generatedStr)))
	}
	return url.QueryEscape(generatedStr)
}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	cryptoRand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func deletion(str string, n int) string {
	if n >= len(str) {
		return ""
	}
	s := strings.Split(str, "")
	for i := 0; i < n; i++ {
		idx := rand.Int() % len(s)
		s = remove(s, idx)
	}
	return strings.Join(s[:], "")
}

func insertion(str string, n int) string {
	s := strings.Split(str, "")
	for i := 0; i < n; i++ {
		idx := rand.Int() % len(s)
		newBase := dna[rand.Int()%len(dna)]
		s = insert(s, idx, newBase)
	}
	return strings.Join(s[:], "")
}

func substitute(str string, n int) string {
	var idxs []int
	s := strings.Split(str, "")
	for i := 0; i < n; i++ {
		idxs = append(idxs, rand.Int()%len(s))
	}
	for idx := range idxs {
		dnaCopy := removeByValue(dna, s[idx])
		newBase := dnaCopy[rand.Int()%len(dnaCopy)]
		s[idx] = newBase
	}
	return strings.Join(s[:], "")
}

func insert(a []string, index int, value string) []string {
	if len(a) == index {
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...)
	a[index] = value
	return a
}

func remove(s []string, index int) []string {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

func removeByValue(s []string, r string) []string {
	strArrCopy := make([]string, len(s))
	j := 0
	for _, v := range s {
		if v != r {
			strArrCopy[j] = v
			j++
		}
	}
	return strArrCopy
}
