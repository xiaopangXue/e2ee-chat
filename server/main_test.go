package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCodeRoomRateLimitByRemoteIP(t *testing.T) {
	h := newHub()
	h.powDifficulty = 8
	for i := 0; i < codeLimit; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/code-room", strings.NewReader(`{"pow":`+mustProofJSON(t, h, "203.0.113.10")+`}`))
		req.RemoteAddr = "203.0.113.10:12345"
		rec := httptest.NewRecorder()
		h.codeRoomHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/code-room", strings.NewReader(`{"pow":`+mustProofJSON(t, h, "203.0.113.10")+`}`))
	req.RemoteAddr = "203.0.113.10:12345"
	rec := httptest.NewRecorder()
	h.codeRoomHandler(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("limited status = %d, want 429", rec.Code)
	}
}

func TestCodeRoomRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	h := newHub()
	h.powDifficulty = 8
	for i := 0; i < codeLimit; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/code-room", strings.NewReader(`{"pow":`+mustProofJSON(t, h, "203.0.113.10")+`}`))
		req.RemoteAddr = "203.0.113.10:12345"
		req.Header.Set("X-Forwarded-For", "198.51.100.1")
		rec := httptest.NewRecorder()
		h.codeRoomHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/code-room", strings.NewReader(`{"pow":`+mustProofJSON(t, h, "203.0.113.10")+`}`))
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.2")
	rec := httptest.NewRecorder()
	h.codeRoomHandler(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("spoofed forwarded-for status = %d, want 429", rec.Code)
	}
}

func TestCodeRoomRateLimitUsesForwardedForFromTrustedProxy(t *testing.T) {
	h := newHub()
	h.powDifficulty = 8
	h.trustedProxy = parseCIDRList("10.0.0.0/8")

	for i := 0; i < codeLimit; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/code-room", strings.NewReader(`{"pow":`+mustProofJSON(t, h, "198.51.100.1")+`}`))
		req.RemoteAddr = "10.0.0.5:12345"
		req.Header.Set("X-Forwarded-For", "198.51.100.1")
		rec := httptest.NewRecorder()
		h.codeRoomHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/code-room", strings.NewReader(`{"pow":`+mustProofJSON(t, h, "198.51.100.2")+`}`))
	req.RemoteAddr = "10.0.0.5:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.2")
	rec := httptest.NewRecorder()
	h.codeRoomHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("different forwarded client status = %d, want 200", rec.Code)
	}
}

func TestJoinCodeRoomRateLimited(t *testing.T) {
	h := newHub()
	h.powDifficulty = 8
	for i := 0; i < codeLimit; i++ {
		req := httptest.NewRequest(http.MethodPut, "/api/code-room", strings.NewReader(`{"code":"123456","pow":`+mustProofJSON(t, h, "203.0.113.20")+`}`))
		req.RemoteAddr = "203.0.113.20:12345"
		rec := httptest.NewRecorder()
		h.codeRoomHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("join request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPut, "/api/code-room", strings.NewReader(`{"code":"123456","pow":`+mustProofJSON(t, h, "203.0.113.20")+`}`))
	req.RemoteAddr = "203.0.113.20:12345"
	rec := httptest.NewRecorder()
	h.codeRoomHandler(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("join limited status = %d, want 429", rec.Code)
	}
}

func TestJoinCodeRoomAcceptsCustomCode(t *testing.T) {
	h := newHub()
	h.powDifficulty = 8
	req := httptest.NewRequest(http.MethodPut, "/api/code-room", strings.NewReader(`{"code":"team-r29","pow":`+mustProofJSON(t, h, "203.0.113.30")+`}`))
	req.RemoteAddr = "203.0.113.30:12345"
	rec := httptest.NewRecorder()
	h.codeRoomHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("custom code status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "/r/TEAMR29#p=TEAMR29") {
		t.Fatalf("custom code response did not normalize code: %s", rec.Body.String())
	}
}

func TestJoinCodeRoomRejectsAmbiguousCode(t *testing.T) {
	h := newHub()
	h.powDifficulty = 8
	req := httptest.NewRequest(http.MethodPut, "/api/code-room", strings.NewReader(`{"code":"ROOM01","pow":`+mustProofJSON(t, h, "203.0.113.31")+`}`))
	req.RemoteAddr = "203.0.113.31:12345"
	rec := httptest.NewRecorder()
	h.codeRoomHandler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("ambiguous code status = %d, want 400", rec.Code)
	}
}

func mustProofJSON(t *testing.T, h *Hub, ip string) string {
	t.Helper()
	challenge, payload, err := h.newPowChallenge(ip, "code")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; ; i++ {
		solution := fmt.Sprintf("test_%d", i)
		hash := sha256.Sum256([]byte(challenge + ":" + solution))
		if hasLeadingZeroBits(hash[:], payload.Difficulty) {
			body, err := json.Marshal(powProof{Challenge: challenge, Solution: solution})
			if err != nil {
				t.Fatal(err)
			}
			return string(body)
		}
	}
}
