package traefik_headers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	headersrwr "github.com/wbpaygate/traefik-headers"
)

type testdata struct {
	head map[string][]string
}

func Test_Headers2(t *testing.T) {
	cases := []struct {
		name  string
		conf  string
		tests []testdata
	}{
		{
			name: "t1",
			conf: `{
  "Content-type1": [
   "connect-src *; frame-ancestors wildberries.ru *.wildberries.ru wildberries.am *.wildberries.am wildberries.kg *.wildberries.kg wildberries.by *.wildberries.by wildberries.kz *.wildberries.kz wildberries.ua *.wildberries.ua wildberries.eu *.wildberries.eu wildberries.ge *.wildberries.ge",
   "aaa"
  ]

}`,
			tests: []testdata{
				{
					head: map[string][]string{
						"cc-bb": {"asdfgh"},
					},
				},
			},
		},
	}

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		io.WriteString(rw, "<html><body>Hello World!</body></html>")
		rw.Header().Set("Content-type", "application/json")
		rw.WriteHeader(400)
	})

	cfg := headersrwr.CreateConfig()
	cfg.HeadersData = `{}`
	_, err := headersrwr.New(context.Background(), next, cfg, "headers")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var tst interface{}
			if err := json.Unmarshal([]byte(tc.conf), &tst); err != nil {
				t.Fatal("init json:", err)
			}
			cfg.HeadersData = tc.conf
			h, err := headersrwr.New(context.Background(), next, cfg, "headers")
			if err != nil {
				t.Fatal(err)
			}

			for _, d := range tc.tests {
				req, err := prepreq("http://aa.vv", d.head)
				if err != nil {
					panic(err)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				for k, v := range rec.HeaderMap {
					fmt.Println("resp h", k, v)
				}

				fmt.Println("code", rec.Code)
				b, _ := io.ReadAll(rec.Body)

				fmt.Println("body", string(b))

				/*
					if rec.Code != 200 {
						t.Errorf("first %s %v expected 200 but get %d", d.uri, d.head, rec.Code)
					}
				*/
			}
		})
	}
}

func prepreq(uri string, head map[string][]string) (*http.Request, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	if head != nil {
		for k, vv := range head {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}
	return req, nil
}
