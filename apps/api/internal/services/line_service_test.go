package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
)

func TestLinkDashboardRichMenu(t *testing.T) {
	requestPath := ""
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestPath = request.URL.Path
		if got := request.Header.Get("Authorization"); got != "Bearer channel-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := &LineService{
		apiBaseURL: server.URL,
		config: config.LineConfig{
			ChannelAccessToken:  "channel-token",
			DashboardRichMenuID: "richmenu-dashboard",
		},
		http: server.Client(),
	}

	if err := service.LinkDashboardRichMenu(context.Background(), "U123"); err != nil {
		t.Fatalf("LinkDashboardRichMenu returned an error: %v", err)
	}
	if requestPath != "/v2/bot/user/U123/richmenu/richmenu-dashboard" {
		t.Fatalf("unexpected request path: %q", requestPath)
	}
}
