package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	http.HandleFunc("/", setup(NewMatcher(getEnv("CONFIG_PATH", "./routingConfig.yaml"))))
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getListenAddress() string {
	port := getEnv("PORT", "1338")
	return ":" + port
}

func setup(matcher Matcher) func(res http.ResponseWriter, req *http.Request) {
	filter := NewSecurityFilter(getEnv("JWK_URL", ""))
	return func(res http.ResponseWriter, req *http.Request) {
		result := matcher.Match(req)
		log.Printf("Matching: %v", result)
		if result == nil {
			res.WriteHeader(404)
		} else {
			if result.Secured {
				filter.Filter(res, req, func() { proxy(result.URL, res, req) })
			} else {
				proxy(result.URL, res, req)
			}
		}
	}
}

func proxy(urlString string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(urlString)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
