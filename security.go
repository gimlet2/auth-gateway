package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"

	"github.com/lestrrat-go/jwx/jwk"
)

type SecurityFilter interface {
	Filter(res http.ResponseWriter, req *http.Request, action func())
}

type SecurityFilterImpl struct {
	pkey jwk.RSAPublicKey
}

func NewSecurityFilter(jwkUrl string) SecurityFilter {
	ks, e := jwk.Fetch(jwkUrl)
	if e != nil {
		log.Fatalf("Faild to load JWK: %v", e)
	}
	key, e := ks.Keys[0].Materialize()
	if e != nil {
		log.Fatalf("Failed to fetch JWK key: %v", e)
	}
	return &SecurityFilterImpl{key.(jwk.RSAPublicKey)}
}

func (s *SecurityFilterImpl) Filter(res http.ResponseWriter, req *http.Request, action func()) {
	_, result := jws.Verify([]byte(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")), jwa.RS256, s.pkey)
	if result != nil {
		res.WriteHeader(401)
		return
	}
	action()
}
