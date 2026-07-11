package services

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestAuthorizationURLContainsOAuthParameters(t *testing.T) {
	service := NewFacebookService(config.FacebookConfig{
		AppID:        "app-id",
		AppSecret:    "app-secret",
		AppURL:       "https://linora.example",
		GraphVersion: "v24.0",
		RedirectURI:  "https://api.linora.example/api/facebook/callback",
	})

	parsed, err := url.Parse(service.AuthorizationURL("state-token"))
	if err != nil {
		t.Fatalf("AuthorizationURL returned invalid URL: %v", err)
	}

	if parsed.Host != "www.facebook.com" || parsed.Path != "/v24.0/dialog/oauth" {
		t.Fatalf("authorization endpoint = %s, want Facebook OAuth endpoint", parsed.String())
	}
	query := parsed.Query()
	if query.Get("client_id") != "app-id" {
		t.Fatalf("client_id = %q, want app-id", query.Get("client_id"))
	}
	if query.Get("redirect_uri") != "https://api.linora.example/api/facebook/callback" {
		t.Fatalf("redirect_uri = %q", query.Get("redirect_uri"))
	}
	if query.Get("state") != "state-token" {
		t.Fatalf("state = %q, want state-token", query.Get("state"))
	}
	if query.Get("scope") != "pages_show_list,pages_read_engagement,pages_read_user_content,read_insights" {
		t.Fatalf("scope = %q", query.Get("scope"))
	}
}

func TestCompleteLoginExchangesCodeAndConsumesSelectedPageOnce(t *testing.T) {
	service := NewFacebookService(config.FacebookConfig{
		AppID:        "app-id",
		AppSecret:    "app-secret",
		AppURL:       "https://linora.example",
		GraphVersion: "v24.0",
		RedirectURI:  "https://api.linora.example/api/facebook/callback",
	})
	service.http = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := ""
		switch request.URL.Path {
		case "/v24.0/oauth/access_token":
			if request.URL.Query().Get("code") != "authorization-code" {
				t.Fatalf("code = %q, want authorization-code", request.URL.Query().Get("code"))
			}
			body = `{"access_token":"user-access-token"}`
		case "/v24.0/me/accounts":
			if request.URL.Query().Get("access_token") != "user-access-token" {
				t.Fatalf("access_token = %q, want user-access-token", request.URL.Query().Get("access_token"))
			}
			body = `{"data":[{"id":"page-1","name":"Linora Cafe","category":"Local business","access_token":"page-token-1"},{"id":"page-2","name":"Linora Studio","category":"Creator","access_token":"page-token-2"}]}`
		default:
			t.Fatalf("unexpected Facebook request: %s", request.URL.String())
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})}

	handoff, err := service.CompleteLogin(context.Background(), "authorization-code", "line-user-1")
	if err != nil {
		t.Fatalf("CompleteLogin returned error: %v", err)
	}

	pages, err := service.RedeemHandoff(handoff, "line-user-1")
	if err != nil {
		t.Fatalf("RedeemHandoff returned error: %v", err)
	}
	if len(pages) != 2 || pages[0].PageID != "page-1" || pages[1].PageName != "Linora Studio" {
		t.Fatalf("pages = %#v, want Facebook page summaries", pages)
	}
	page, err := service.ConsumePage(handoff, "line-user-1", "page-1")
	if err != nil || page.AccessToken != "page-token-1" {
		t.Fatalf("ConsumePage = %#v, %v", page, err)
	}
	if _, err := service.ConsumePage(handoff, "line-user-1", "page-1"); err == nil {
		t.Fatal("ConsumePage accepted a handoff code twice")
	}
}
