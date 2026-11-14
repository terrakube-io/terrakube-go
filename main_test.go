package terrakube

import (
	"net/http"
	"os"
	"testing"
)

// TestMain gives the tests their own HTTP connection pool. A client built
// without WithHTTPClient uses http.DefaultClient, and every httptest server
// closes the idle connections of http.DefaultTransport when it shuts down.
// With hundreds of parallel tests that sometimes hits a connection another
// test has just picked for a POST, PATCH or DELETE, which net/http does not
// retry: "http: CloseIdleConnections called".
func TestMain(m *testing.M) {
	http.DefaultClient = &http.Client{Transport: http.DefaultTransport.(*http.Transport).Clone()}
	os.Exit(m.Run())
}
