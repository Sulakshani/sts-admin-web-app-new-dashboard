package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sts-backend/internal/config"
)

func TestSendDialogSMSAPIv2_Success(t *testing.T) {
	sawSMSRequest := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			t.Fatal("login endpoint should not be called when DialogSMSAccessToken is configured")
		}
		if strings.HasSuffix(r.URL.Path, "/sms") {
			sawSMSRequest = true
			if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
				t.Fatalf("Authorization header = %q, want %q", got, want)
			}
			if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
				t.Fatalf("Content-Type header = %q, want %q", got, want)
			}

			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode sms payload: %v", err)
			}
			if got, want := payload["message"], "Hello Test"; got != want {
				t.Fatalf("message = %q, want %q", got, want)
			}
			if got, want := payload["sourceAddress"], "TestMask"; got != want {
				t.Fatalf("sourceAddress = %q, want %q", got, want)
			}
			if got, want := payload["payment_method"], float64(0); got != want {
				t.Fatalf("payment_method = %v, want %v", got, want)
			}
			if got := payload["transaction_id"]; got == nil {
				t.Fatal("transaction_id is missing")
			}

			msisdn, ok := payload["msisdn"].([]interface{})
			if !ok || len(msisdn) != 1 {
				t.Fatalf("msisdn = %#v, want one recipient", payload["msisdn"])
			}
			firstRecipient, ok := msisdn[0].(map[string]interface{})
			if !ok {
				t.Fatalf("msisdn[0] = %#v, want object", msisdn[0])
			}
			if got, want := firstRecipient["mobile"], "712345678"; got != want {
				t.Fatalf("mobile = %q, want %q", got, want)
			}

			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"status":"success","comment":"Campaign Created","data":{"campaignId":12345},"errCode":""}`)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	cfg := &config.Config{
		DialogSMSMethod:      "api_v2",
		DialogSMSAPIURL:      server.URL,
		DialogSMSAccessToken: "test-token",
		DialogSMSMask:        "TestMask",
	}

	smsService := NewSMSService(cfg)

	err := smsService.sendDialogSMSAPIv2("94712345678", "Hello Test")
	if err != nil {
		t.Fatalf("sendDialogSMSAPIv2 failed: %v", err)
	}
	if !sawSMSRequest {
		t.Fatal("expected SMS request to be sent")
	}
}

func TestSendDialogSMSAPIv2_LoginFallbackSuccess(t *testing.T) {
	sawLoginRequest := false
	sawSMSRequest := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			sawLoginRequest = true
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"status":"success","comment":"SUCCESS","accessToken":"test-token","tokenValidity":"3600","errCode":""}`)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/sms") {
			sawSMSRequest = true
			if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
				t.Fatalf("Authorization header = %q, want %q", got, want)
			}
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"status":"success","comment":"Campaign Created","data":{"campaignId":12345},"errCode":""}`)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	cfg := &config.Config{
		DialogSMSMethod:   "api_v2",
		DialogSMSAPIURL:   server.URL,
		DialogSMSUsername: "testuser",
		DialogSMSPassword: "testpass",
		DialogSMSMask:     "TestMask",
	}

	smsService := NewSMSService(cfg)

	if err := smsService.sendDialogSMSAPIv2("94712345678", "Hello Test"); err != nil {
		t.Fatalf("sendDialogSMSAPIv2 failed: %v", err)
	}
	if !sawLoginRequest {
		t.Fatal("expected login fallback request to be sent")
	}
	if !sawSMSRequest {
		t.Fatal("expected SMS request to be sent")
	}
}

func TestNormalizeDialogAPIv2Mobile(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "country code", in: "94715342627", want: "715342627"},
		{name: "local leading zero", in: "0715342627", want: "715342627"},
		{name: "already api format", in: "715342627", want: "715342627"},
		{name: "with formatting", in: "+94 71 534 2627", want: "715342627"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDialogAPIv2Mobile(tt.in)
			if err != nil {
				t.Fatalf("normalizeDialogAPIv2Mobile(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeDialogAPIv2Mobile(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeDialogAPIv2MobileRejectsInvalidNumber(t *testing.T) {
	if _, err := normalizeDialogAPIv2Mobile("12345"); err == nil {
		t.Fatal("expected invalid number to return an error")
	}
}

func TestSendDialogSMSAPIv2_NumericErrorCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sms") {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, `{"status":"failed","comment":"Invalid token signature detected","data":"","errCode":105,"transaction_id":null}`)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	cfg := &config.Config{
		DialogSMSMethod:      "api_v2",
		DialogSMSAPIURL:      server.URL,
		DialogSMSAccessToken: "invalid-token",
		DialogSMSMask:        "TestMask",
	}

	err := NewSMSService(cfg).sendDialogSMSAPIv2("94712345678", "Hello Test")
	if err == nil {
		t.Fatal("sendDialogSMSAPIv2 should have failed but did not")
	}

	expectedError := "Dialog SMS API error: Invalid token signature detected (Code: 105)"
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("Expected error to contain %q, but got %q", expectedError, err.Error())
	}
}

func TestSendDialogSMSAPIv2_LoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, `{"status":"failed","comment":"Invalid credentials","errCode":"E1001"}`)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	cfg := &config.Config{
		DialogSMSMethod:   "api_v2",
		DialogSMSAPIURL:   server.URL,
		DialogSMSUsername: "wronguser",
		DialogSMSPassword: "wrongpassword",
	}

	smsService := NewSMSService(cfg)

	err := smsService.sendDialogSMSAPIv2("94712345678", "Hello Test")
	if err == nil {
		t.Fatal("sendDialogSMSAPIv2 should have failed but did not")
	}

	expectedError := "Dialog API login failed: Invalid credentials (Code: E1001)"
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("Expected error to contain %q, but got %q", expectedError, err.Error())
	}
}
