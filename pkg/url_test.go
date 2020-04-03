package urlshortener

import (
	"strings"
	"testing"
)

// NormalizeURL checks if the given string is a valid url
// and prepends http:// if it doesn't start with 'http://' or 'https://'
func TestURLCantBeEmpty(t *testing.T) {
	_, err := NormalizeURL("")
	if err == nil {
		t.Fatal("Empty url should fail validation")
	}
}
func TestURLCantBeLargerThan2048Characters(t *testing.T) {
	url := "https://www.example.com/param?query="
	url = url + stringGen(2049-len(url), 'a')
	_, err := NormalizeURL(url)
	if err == nil {
		t.Fatal("A 2049 character url should fail validation")
	}
}

func TestURLShouldBeValid(t *testing.T) {
	cases := []string{
		`javascript:alert("Hello World");`,
		"",
		"   ",
	}
	for _, url := range cases {
		_, err := NormalizeURL(url)
		if err == nil {
			t.Fatal("Bad url should fail validation:", url)
		}
	}
}

func TestURLShouldBeAppendedWithHttpScheme(t *testing.T) {
	url := "www.example.com"
	res, err := NormalizeURL(url)
	if err != nil {
		t.Fatal("Good URL failed validation:", url)
	}
	if strings.HasPrefix(res, "http://") == false {
		t.Fatal("Url should start with http", res)
	}
}

func TestGoodUrlsPass(t *testing.T) {
	example := "https://www.example.com/param?query="
	maxLengthURL := example + stringGen(2048-len(example), 'a')
	cases := []string{
		example,
		maxLengthURL,
		"example.com",
		"www.example.com",
		"sub.sub.sub.domain.com",
		"http://www.example.com/",
		"https://www.example.com/?query=30#test",
	}
	for _, url := range cases {
		_, err := NormalizeURL(url)
		if err != nil {
			t.Fatal("Good url should pass validation:", url)
		}
	}
}

func stringGen(size int, char rune) string {
	query := make([]rune, size)
	for i := 0; i < size; i++ {
		query[i] = 'a'
	}
	return string(query)
}
