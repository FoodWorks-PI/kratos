package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/ory/herodot"
	"github.com/ory/x/stringsx"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	defaultBaseUrl = "https://graph.facebook.com/v8.0"
)

type ProviderFacebook struct {
	config *Configuration
	public *url.URL
}

func NewProviderFacebook(
	config *Configuration,
	public *url.URL,
) *ProviderFacebook {
	config.IssuerURL = "https://graph.facebook.com/v8.0"
	return &ProviderFacebook{
		config: config,
		public: public,
	}
}

func (f *ProviderFacebook) Config() *Configuration {
	return f.config
}

func (f *ProviderFacebook) oauth2() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     f.config.ClientID,
		ClientSecret: f.config.ClientSecret,
		Endpoint:     f.endpoint(),
		Scopes:       f.config.Scope,
		RedirectURL:  f.config.Redir(f.public),
	}
}

func (g *ProviderFacebook) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(), nil
}

func (g *ProviderFacebook) AuthCodeURLOptions(r request) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (f *ProviderFacebook) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	log.Print(exchange.AccessToken)

	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), ",")
	log.Print(grantedScopes)
	for _, scope := range grantedScopes {
		log.Print(scope)
	}
	grantedScopes = stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scopes")), ",")
	log.Print(grantedScopes)
	for _, scope := range grantedScopes {
		log.Print(scope)
	}
	/*
		for _, check := range f.Config().Scope {
			if !stringslice.Has(grantedScopes, check) {
				log.Print(check)
				return nil, errors.WithStack(ErrScopeMissing)
			}
		}
	*/

	client := f.oauth2().Client(ctx, exchange)
	u, err := url.Parse(f.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	u.Path = path.Join(u.Path, "/me")
	q := u.Query()
	q.Set("fields", "id,name,email")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	log.Println("REQUESTING")
	log.Println(req.RequestURI)
	log.Println(u.Path)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed")
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	defer resp.Body.Close()

	var claims Claims
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Print(bodyString)
	err = json.Unmarshal(bodyBytes, &claims)
	if err != nil {
		log.Println("Failed decoding")
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	/*
		if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
			log.Println("Failed decoding")
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}
	*/
	log.Print(claims)
	return &claims, nil
}

func (f *ProviderFacebook) endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  f.config.AuthURL,
		TokenURL: f.config.TokenURL,
	}
}
