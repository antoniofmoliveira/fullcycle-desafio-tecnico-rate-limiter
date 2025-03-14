package main

import (
	"net/http"
	"net/url"
	"sync"
	"testing"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name  string
		token string
		qtReq int
		want  map[int]int
	}{
		{
			name:  "No token",
			token: "",
			qtReq: 500,
			want:  map[int]int{429: 495, 200: 5},
		},
		{
			name:  "Token level 1",
			token: "1234",
			qtReq: 500,
			want:  map[int]int{429: 490, 200: 10},
		},
		{
			name:  "Token level 2",
			token: "2234",
			qtReq: 500,
			want:  map[int]int{429: 480, 200: 20},
		},
		{
			name:  "Token level 3",
			token: "3234",
			qtReq: 500,
			want:  map[int]int{429: 450, 200: 50},
		},
		{
			name:  "Token level 4",
			token: "4234",
			qtReq: 500,
			want:  map[int]int{429: 400, 200: 100},
		},
		{
			name:  "Token fail",
			token: "2234",
			qtReq: 500,
			want:  map[int]int{429: 500},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := execute(tt.token, tt.qtReq)

			if len(got) != len(tt.want) {
				t.Errorf("main() = %v, want %v", got, tt.want)
			}

			for k, v := range got {
				if v != tt.want[k] {
					t.Errorf("status %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func execute(token string, qtReq int) map[int]int {
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
	return results
}
