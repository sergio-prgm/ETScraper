package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Arr []string

type Map map[string]Arr

var examCodes = Map{
	"AZ-104":              make([]string, 0),
	"AZ-204":              make([]string, 0),
	"Terraform Associate": make([]string, 0),
	"DP-100":              make([]string, 0),
	"AZ-400":              make([]string, 0),
}

func (arr Arr) Print() string {
	result := strings.Join(arr, ",")
	return result
}

func SaveCodes(emap Map) {
	filename := ""
	filecontent := ""
	filename = "codes.txt"

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	for key, arr := range emap {
		filecontent += fmt.Sprintf("#%s\n%s\n", key, arr.Print())
	}

	w := bufio.NewWriter(f)

	_, err = w.WriteString(filecontent)

	if err != nil {
		panic(err)
	}
	w.Flush()
}

func (emap Map) Contains(exam string) bool {
	result := false
	for key := range emap {
		result = key == exam
		if result {
			return result
		}
	}
	return result
}

func randomSleep(start int) {
	seconds := rand.Intn(20) + start
	fmt.Println(seconds)
	time.Sleep(time.Second * time.Duration(seconds))
}

func parseHeaders() map[string]string {
	hFile, err := os.Open("headers.json")
	if err != nil {
		fmt.Println("Something went wrong while reading the file")
	}
	defer hFile.Close()
	var headers map[string]string

	body, _ := io.ReadAll(hFile)

	// TODO Probably should handle errors somehow
	json.Unmarshal(body, &headers)
	if err != nil {
		fmt.Print("Error during unmarshaling")
	}
	return headers
}

func main() {
	initialCode := 56828
	// initialCode := 57038
	// code := 79381 Terraform associate
	maxCodes := 200
	headers := parseHeaders()
	// https://www.scraperapi.com/blog/10-tips-for-web-scraping/
	for code := initialCode; code < initialCode+maxCodes; code++ {
		url := fmt.Sprintf("https://www.examtopics.com/discussions/microsoft/view/%d-exam/", code)

		// res, _ := http.Get(url)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			fmt.Print("Error making request")
			os.Exit(1)
		}

		for key, val := range headers {
			// TODO rotate user-agents (maybe other keys as well)
			// gDesktopAgent := Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36
			// gMobileAgent := Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)
			req.Header.Set(key, val)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Error making http request to %s", url)
			os.Exit(1)
		}

		fmt.Print("\n")
		exam := parse(res.Body)
		fmt.Printf("%d - %s\n", code, exam)

		if !examCodes.Contains(exam) {
			randomSleep(15)
			continue
		}

		var currentCodes Arr = examCodes[exam]
		examCodes[exam] = append(currentCodes, fmt.Sprintf("%d", code))

		fmt.Println(examCodes)
		randomSleep(22)
		defer res.Body.Close()
	}
	SaveCodes(examCodes)
}

func parse(text io.Reader) (data string) {
	z := html.NewTokenizer(text)
	var isH1 bool
	var exam string

	for {
		tt := z.Next()
		switch tt {

		case html.ErrorToken:
			return exam
		case html.TextToken:
			t := z.Token().Data
			// fmt.Print(t)
			if isH1 {
				// fmt.Print(t)
				isH1 = false
				tarr := strings.Split(t, " ")
				exam = strings.Join(tarr[1:len(tarr)-5], " ")
			}
		case html.StartTagToken:
			if isH1 {
				continue
			}
			tag := z.Token().Data
			isH1 = tag == "h1"
		}
	}
}
