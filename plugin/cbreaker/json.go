package cbreaker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mailgun/vulcand/Godeps/_workspace/src/github.com/mailgun/vulcan/circuitbreaker"
	"github.com/mailgun/vulcand/Godeps/_workspace/src/github.com/mailgun/vulcan/middleware"
)

type rawAction struct {
	Type   string
	Action json.RawMessage
}

type rawResponse struct {
	StatusCode  int
	ContentType string
	Body        json.RawMessage
}

type rawSideEffect struct {
	Type   string
	Action json.RawMessage
}

type rawWebhook struct {
	URL     string
	Method  string
	Headers http.Header
	Form    url.Values
	Body    json.RawMessage
}

func actionFromJSON(v []byte) (middleware.Middleware, error) {
	var a *rawAction
	err := json.Unmarshal(v, &a)
	if err != nil {
		return nil, err
	}
	switch a.Type {
	case "redirect":
		return redirectFromJSON(a.Action)
	case "response":
		return responseFromJSON(a.Action)
	}
	return nil, fmt.Errorf("unsupported action: '%s' expected 'redirect' or 'reply'", a.Action)
}

func sideEffectFromJSON(v []byte) (circuitbreaker.SideEffect, error) {
	var a *rawAction
	err := json.Unmarshal(v, &a)
	if err != nil {
		return nil, err
	}
	switch a.Type {
	case "webhook":
		return webhookFromJSON(a.Action)
	}
	return nil, fmt.Errorf("unsupported action: '%s' expected 'webhook'", a.Action)
}

func redirectFromJSON(v []byte) (*circuitbreaker.RedirectFallback, error) {
	var r *circuitbreaker.Redirect
	err := json.Unmarshal(v, &r)
	if err != nil {
		return nil, err
	}
	return circuitbreaker.NewRedirectFallback(*r)
}

func responseFromJSON(v []byte) (*circuitbreaker.ResponseFallback, error) {
	var r *rawResponse
	err := json.Unmarshal(v, &r)
	if err != nil {
		return nil, err
	}
	b, err := responseBodyFromJSON(r.Body)
	if err != nil {
		return nil, err
	}
	return circuitbreaker.NewResponseFallback(
		circuitbreaker.Response{
			StatusCode:  r.StatusCode,
			Body:        b,
			ContentType: r.ContentType,
		})
}

func responseBodyFromJSON(v []byte) ([]byte, error) {
	// Try to decode bytes first (expects base64 encoded array)
	var bytes []byte
	if err := json.Unmarshal(v, &bytes); err == nil {
		return bytes, nil
	}

	// Try to decode string next
	var str string
	if err := json.Unmarshal(v, &str); err == nil {
		return []byte(str), nil
	}

	// In case if it's a valid JSON object, return it as-is
	var obj map[string]interface{}
	if err := json.Unmarshal(v, &obj); err == nil {
		return v, nil
	}

	return nil, fmt.Errorf("expected string, bytes or object")
}

func webhookFromJSON(v []byte) (*circuitbreaker.WebhookSideEffect, error) {
	var w *rawWebhook
	if err := json.Unmarshal(v, &w); err != nil {
		return nil, err
	}

	var body []byte
	var err error
	if len(w.Body) != 0 {
		body, err = responseBodyFromJSON(w.Body)
		if err != nil {
			return nil, err
		}
	}

	return circuitbreaker.NewWebhookSideEffect(
		circuitbreaker.Webhook{
			URL:     w.URL,
			Method:  w.Method,
			Headers: w.Headers,
			Form:    w.Form,
			Body:    body,
		})
}
