package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

type Matcher interface {
	Match(req *http.Request) *Pair
}

type MatcherImpl struct {
	paths map[string]map[string]map[string]Pair
}

func NewMatcher(path string) Matcher {
	proxy, err := readConfig(path)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	var config = make(map[string]map[string]map[string]Pair)
	for _, host := range proxy.Hosts {
		for _, route := range host.Routes {
			_, p := config[route.Path]
			if p == false {
				config[route.Path] = make(map[string]map[string]Pair)
			}
			_, p = config[route.Path][route.HTTPMethod]
			if p == false {
				config[route.Path][route.HTTPMethod] = make(map[string]Pair)
			}
			config[route.Path][route.HTTPMethod][route.Version] = Pair{route.Secured, host.URL}
			log.Printf("Add route: %s, %s, %s, %s", route.Path, route.HTTPMethod, route.Version, host.URL)
		}
	}
	return &MatcherImpl{config}
}

func (m *MatcherImpl) Match(req *http.Request) *Pair {
	v1, p := m.paths[req.URL.Path]
	if p == false {
		return nil
	}
	v2, p := v1[req.Method]
	if p == false {
		return nil
	}
	version, pv := req.Header["Accepted-Version"]
	var v3 Pair
	if pv == false {
		v3, p = v2[""]
		if p == false {
			return nil
		}
	} else {
		v3, p = v2[version[0]]
		if p == false {
			return nil
		}
	}
	return &v3
}

func readConfig(path string) (*Proxy, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read proxt config: %v", err)
	}
	var config ProxyRoot
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal proxy: %v", err)
	}
	return &config.Proxy, nil
}

type ProxyRoot struct {
	Proxy Proxy `yaml:"proxy"`
}

type Proxy struct {
	Hosts []Host `yaml:"hosts"`
}

type Host struct {
	URL    string  `yaml:"url"`
	Routes []Route `yaml:"routes"`
}

type Route struct {
	HTTPMethod string `yaml:"httpMethod"`
	Path       string `yaml:"path"`
	Version    string `yaml:"version"`
	Secured    bool   `yaml:"secured"`
}

type Pair struct {
	Secured bool
	URL     string
}
