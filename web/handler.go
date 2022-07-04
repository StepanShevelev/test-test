package web

import (
	"context"
	"encoding/json"
	"github.com/StepanShevelev/test-test/config"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type input struct {
	URL []string `json:"urls"`
}

type Response struct {
	URL        string            `json:"url"`
	StatusCode int               `json:"status_code"`
	Response   string            `json:"response"`
	Headers    map[string]string `json:"headers"`
}

type successResponse struct {
	OK     bool       `json:"ok"`
	Result []Response `json:"result"`
}

func Init() http.Handler {
	return http.HandlerFunc(handle)
}

func handle(w http.ResponseWriter, r *http.Request) {

	if !isMethodPOST(w, r) {
		return
	}

	cfg := config.New()
	if err := cfg.Load("./configs", "config", "yml"); err != nil {
		log.Fatal(err)
	}
	var limitConn = make(chan struct{}, cfg.Limit)

	limitConn <- struct{}{}
	defer func() { <-limitConn }()

	log.Printf("%s request: %s", r.Method, r.URL)

	handleUrls(w, r)

}

func handleUrls(w http.ResponseWriter, r *http.Request) {
	i := input{}

	cfg := config.New()
	if err := cfg.Load("./configs", "config", "yml"); err != nil {
		log.Fatal(err)
	}

	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON"))
		return
	}

	switch {
	case len(i.URL) == 0:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No URL`s to process"))
		return
	case len(i.URL) > 20:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Can`t process more than 20 URL`s"))
		return
	}

	// create context with cancel to stop all goroutines
	ctx, cancel := context.WithCancel(r.Context())

	// 4 - limit
	ch := make(chan Response, cfg.Req)

	defer cancel()

	var wg sync.WaitGroup

	for _, u := range i.URL {
		_, err := url.ParseRequestURI(u)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid URL"))
			return
		}

		wg.Add(1)
		go func(ctx context.Context, url string) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					ch <- Response{URL: url}
					log.Fatalf("Error creating request: %v", err)
				}
				req = req.WithContext(ctx)

				client := http.Client{Timeout: 1 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					ch <- Response{URL: url}
					return
				}
				defer func() {
					err = resp.Body.Close()
					if err != nil {
						log.Printf("Error closing body : %v", err.Error())
					}
				}()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Error read body; %v", err)
				}

				headers := make(map[string]string)
				getHeaders(resp.Header, headers)

				ch <- Response{
					URL:        url,
					StatusCode: resp.StatusCode,
					Response:   string(body),
					Headers:    headers,
				}
			}
		}(ctx, u)
	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	var res []Response
	for ur := range ch {

		if ur.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error handling URL"))
			w.Write([]byte(ur.URL))
			return
		}

		// write result in urlResponse slice
		res = append(res, ur)
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(&successResponse{
		OK:     true,
		Result: res,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func isMethodPOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func getHeaders(in http.Header, out map[string]string) {
	for k, v := range in {
		out[k] = v[0]
	}
}


