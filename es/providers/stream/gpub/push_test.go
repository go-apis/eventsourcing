package gpub

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-apis/eventsourcing/es"
)

func newTestPushStreamer() *pushStreamer {
	return &pushStreamer{
		service:  "fraud",
		errCh:    make(chan error, 10),
		handlers: map[string]es.MessageHandler{},
	}
}

func envelope(subscription, data string) string {
	return fmt.Sprintf(`{
		"message": {"data": %q, "messageId": "m1"},
		"subscription": %q
	}`, base64.StdEncoding.EncodeToString([]byte(data)), subscription)
}

func post(h http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/pubsub/push", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestPushHandlerAcks(t *testing.T) {
	s := newTestPushStreamer()

	var got []byte
	s.handlers["inflow__fraud"] = func(ctx context.Context, payload []byte) error {
		got = payload
		return nil
	}

	w := post(s.PushHandler(), envelope("projects/p/subscriptions/inflow__fraud", "hello"))
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if string(got) != "hello" {
		t.Fatalf("expected decoded payload %q, got %q", "hello", got)
	}
}

func TestPushHandlerNacksOnHandlerError(t *testing.T) {
	s := newTestPushStreamer()
	s.handlers["inflow__fraud"] = func(ctx context.Context, payload []byte) error {
		return errors.New("boom")
	}

	w := post(s.PushHandler(), envelope("projects/p/subscriptions/inflow__fraud", "x"))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 (nack), got %d", w.Code)
	}

	select {
	case err := <-s.Errors():
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
		t.Fatal("expected handler error on error channel")
	}
}

func TestPushHandlerUnknownSubscription(t *testing.T) {
	s := newTestPushStreamer()

	w := post(s.PushHandler(), envelope("projects/p/subscriptions/inflow__other", "x"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPushHandlerBadEnvelope(t *testing.T) {
	s := newTestPushStreamer()

	w := post(s.PushHandler(), "{not json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPushHandlerRejectsGet(t *testing.T) {
	s := newTestPushStreamer()

	req := httptest.NewRequest(http.MethodGet, "/pubsub/push", nil)
	w := httptest.NewRecorder()
	s.PushHandler().ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestAddHandlerNaming(t *testing.T) {
	s := newTestPushStreamer()
	s.topicId = "inflow"

	if err := s.AddHandler(context.Background(), "", func(ctx context.Context, payload []byte) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.handlers["inflow__fraud"]; !ok {
		t.Fatal("expected base subscription registration")
	}

	if err := s.AddHandler(context.Background(), "sub2", func(ctx context.Context, payload []byte) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.handlers["inflow__fraud-sub2"]; !ok {
		t.Fatal("expected suffixed subscription registration")
	}

	if err := s.AddHandler(context.Background(), "", func(ctx context.Context, payload []byte) error { return nil }); err == nil {
		t.Fatal("expected duplicate registration to error")
	}
}
