package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func main() {
	numReqs := 500

	test("No token", "", numReqs)
	test("Token level 1", "1234", numReqs)
	test("Token level 2", "2234", numReqs)
	test("Token level 3", "3234", numReqs)
	test("Token level 4", "4234", numReqs)
	test("Token fail", "5234", numReqs)
}

func test(title string, token string, qtReq int) {
	url, err := url.Parse("http://localhost:8080")
	if err != nil {
		panic(err)
	}
	header := http.Header{}
	if token != "" {
		header.Set("API_KEY", token)
	}

	wg := sync.WaitGroup{}
	wg.Add(qtReq)
	ch := make(chan int, qtReq)
	results := make(map[int]int)
	start := time.Now()

	for range qtReq {
		go func() {
			req, err := http.NewRequest("GET", url.String(), nil)
			if err != nil {
				panic(err)
			}
			req.Header = header
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			ch <- resp.StatusCode
			wg.Done()
		}()
	}

	for range qtReq {
		result := <-ch
		results[result]++
	}
	wg.Wait()
	timeElapsed := time.Since(start)

	fmt.Printf(title + "\n")
	fmt.Printf("Qt requests: %v by second\n", qtReq)
	fmt.Printf("Status code:\tQt Resp\n")
	for k, v := range results {
		fmt.Printf("%v:\t\t%v\n", k, v)
	}
	fmt.Printf("Time elapsed: %v\n\n", timeElapsed)
	time.Sleep(time.Second)
}
