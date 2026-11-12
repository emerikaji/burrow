package network

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverRefs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/info/refs" {
			t.Errorf("expected path /info/refs, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("service") != "git-upload-pack" {
			t.Errorf("expected service=git-upload-pack, got %s", r.URL.Query().Get("service"))
		}

		if ua := r.Header.Get("User-Agent"); ua == "" {
			t.Errorf("expected User-Agent header to be set")
		}

		w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
		w.WriteHeader(http.StatusOK)

		// Stream structure:
		// - Service header packet
		// - Flush packet (0000)
		// - Ref entry for main branch (with capabilities null byte)
		// - Ref entry for a tag
		// - Trailing flush packet (0000)
		mockPktStream :=
			"001e# service=git-upload-pack\n" +
				"0000" +
				"00600123456789abcdef0123456789abcdef01234567 refs/heads/main\x00symref=refs/heads/master thin-pack\n" +
				"003cfedcba9876543210fedcba9876543210fedcba98 refs/tags/v1.0\n" +
				"0000"

		w.Write([]byte(mockPktStream))
	}))
	defer ts.Close()

	refs, caps, err := DiscoverRefs(ts.URL)
	if err != nil {
		t.Fatalf("DiscoverRefs failed: %v", err)
	}

	expectedMainHash := "0123456789abcdef0123456789abcdef01234567"
	if refs["refs/heads/main"] != expectedMainHash {
		t.Errorf("expected main ref %s, got %s", expectedMainHash, refs["refs/heads/main"])
	}

	expectedTagHash := "fedcba9876543210fedcba9876543210fedcba98"
	if refs["refs/tags/v1.0"] != expectedTagHash {
		t.Errorf("expected tag ref %s, got %s", expectedTagHash, refs["refs/tags/v1.0"])
	}

	expectedRef := "refs/heads/master"
	if caps["symref"] != expectedRef {
		t.Errorf("expected symref %s, got %s", expectedRef, caps["symref"])
	}
}

func TestDiscoverRefs_ServerErrors(t *testing.T) {
	// Test behavior when the server returns a 404 or 500 status code
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	_, _, err := DiscoverRefs(ts.URL)
	if err == nil {
		t.Errorf("expected error for 404 response, got nil")
	}
}
