package maas

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/canonical/gomaasclient/client"
)

const testBootResourceID = 42

// newTestMAASClient points a gomaasclient at a stub MAAS API server.
func newTestMAASClient(t *testing.T, handler http.Handler) *client.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	// gomaasclient rejects an API key that is not three colon separated parts.
	c, err := client.GetClient(server.URL, "consumer:token:secret", "2.0")
	if err != nil {
		t.Fatalf("building test client: %v", err)
	}

	return c
}

// serveBootResourceList answers both the list operation and the import
// operation, which share the boot-resources collection endpoint.
func serveBootResourceList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]map[string]any{{
		"name":         "ubuntu/jammy",
		"architecture": "amd64/generic",
		"type":         "synced",
		"id":           testBootResourceID,
	}})
}

// routeBootResource dispatches on the path suffix so the test is unaffected by
// whether the collection endpoint carries a trailing slash. It returns false
// when the request targets neither the collection nor the known resource.
func routeBootResource(w http.ResponseWriter, r *http.Request, detail func(w http.ResponseWriter)) bool {
	index := strings.Index(strings.TrimSuffix(r.URL.Path, "/"), "/boot-resources")
	if index < 0 {
		return false
	}

	rest := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/")[index:], "/boot-resources")
	rest = strings.TrimPrefix(rest, "/")

	switch rest {
	case "":
		serveBootResourceList(w, r)
		return true
	case strconv.Itoa(testBootResourceID):
		detail(w)
		return true
	default:
		return false
	}
}

// A boot resource can be removed and recreated while the import is still
// running, so the per-resource fetch can 404 moments after the resource
// appeared in the list. That has to keep the wait going, not abort it.
func TestAwaitImportCompleteRetriesMissingBootResource(t *testing.T) {
	var fetches atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !routeBootResource(w, r, func(w http.ResponseWriter) {
			if fetches.Add(1) <= 2 {
				// gomaasapi renders this as "ServerError: 404 Not Found (...)".
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, `{"detail": "No BootResource matches the given query."}`)

				return
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"name":"ubuntu/jammy","architecture":"amd64/generic","id":42,`+
				`"sets":{"24.04-ga":{"complete":true,"version":"24.04"}}}`)
		}) {
			http.NotFound(w, r)
		}
	})

	err := awaitImportComplete(
		newTestMAASClient(t, handler),
		"ubuntu", "jammy", []string{"amd64"}, 30*time.Second,
	)
	if err != nil {
		t.Fatalf("awaitImportComplete failed after a transient 404: %v", err)
	}

	if got := fetches.Load(); got != 3 {
		t.Errorf("resource fetched %d times, want 3 (two 404s then success)", got)
	}
}

// Only a 404 means "still importing". Everything else is a real failure and
// must surface straight away rather than burn the whole timeout.
func TestAwaitImportCompleteFailsOnOtherErrors(t *testing.T) {
	var fetches atomic.Int32

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !routeBootResource(w, r, func(w http.ResponseWriter) {
			fetches.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"errors": ["boom"]}`)
		}) {
			http.NotFound(w, r)
		}
	})

	err := awaitImportComplete(
		newTestMAASClient(t, handler),
		"ubuntu", "jammy", []string{"amd64"}, 30*time.Second,
	)
	if err == nil {
		t.Error("awaitImportComplete succeeded, want an error")
	} else if strings.Contains(err.Error(), "404 Not Found") {
		t.Errorf("error should not be a 404: %v", err)
	}

	if got := fetches.Load(); got != 1 {
		t.Errorf("resource fetched %d times, want 1 (a non-404 error must not be retried)", got)
	}
}
