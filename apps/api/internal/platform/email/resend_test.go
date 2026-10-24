package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResendSendsTheExpectedRequest(t *testing.T) {
	var got resendRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer re_test" {
			t.Errorf("Authorization = %q", auth)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	client := NewResend("re_test", "Kyno <kyno@example.com>")
	client.endpoint = server.URL

	err := client.Send(context.Background(), Message{To: "a@example.com", Subject: "Hi", HTML: "<p>Hi</p>", Text: "Hi"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if got.From != "Kyno <kyno@example.com>" || len(got.To) != 1 || got.To[0] != "a@example.com" || got.Subject != "Hi" {
		t.Fatalf("unexpected request body: %+v", got)
	}
}

func TestResendReturnsAnErrorOnFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"invalid from address"}`, http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	client := NewResend("re_test", "bad")
	client.endpoint = server.URL

	err := client.Send(context.Background(), Message{To: "a@example.com"})
	if err == nil || !strings.Contains(err.Error(), "422") {
		t.Fatalf("error = %v; want one mentioning status 422", err)
	}
}
