package blocklist

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// ---- parseBlockList (unexported, tested via package-level access) ----

func TestParseBlockList_EmptyInput(t *testing.T) {
	domains := parseBlockList("")
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}

func TestParseBlockList_SkipsComments(t *testing.T) {
	input := "# this is a comment\n# another comment\n"
	domains := parseBlockList(input)
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d: %v", len(domains), domains)
	}
}

func TestParseBlockList_SkipsEmptyLines(t *testing.T) {
	input := "\n\n   \n\t\n"
	domains := parseBlockList(input)
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}

func TestParseBlockList_PlainDomains(t *testing.T) {
	input := "example.com\nbad.com\n"
	domains := parseBlockList(input)
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d: %v", len(domains), domains)
	}
	if domains[0] != "example.com" {
		t.Errorf("expected example.com, got %q", domains[0])
	}
	if domains[1] != "bad.com" {
		t.Errorf("expected bad.com, got %q", domains[1])
	}
}

func TestParseBlockList_HostsFormat(t *testing.T) {
	// Standard /etc/hosts-style blocklist: "0.0.0.0 domain.com"
	input := "0.0.0.0 ads.example.com\n127.0.0.1 tracker.example.com\n"
	domains := parseBlockList(input)
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d: %v", len(domains), domains)
	}
	if domains[0] != "ads.example.com" {
		t.Errorf("expected ads.example.com, got %q", domains[0])
	}
	if domains[1] != "tracker.example.com" {
		t.Errorf("expected tracker.example.com, got %q", domains[1])
	}
}

func TestParseBlockList_StripsWhitespace(t *testing.T) {
	input := "  example.com  \n\t  bad.com\t\n"
	domains := parseBlockList(input)
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d: %v", len(domains), domains)
	}
}

func TestParseBlockList_MixedContent(t *testing.T) {
	input := `# blocklist
0.0.0.0 ads.com
# another comment

127.0.0.1 tracker.com
plain.com
`
	domains := parseBlockList(input)
	if len(domains) != 3 {
		t.Fatalf("expected 3 domains, got %d: %v", len(domains), domains)
	}
}

// ---- LoadLocalFile ----

func TestLoadLocalFile_NotFound(t *testing.T) {
	_, err := LoadLocalFile("/nonexistent/path/to/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadLocalFile_ValidFile(t *testing.T) {
	content := "example.com\nbad.com\n# comment\n\n"
	f := writeTempFile(t, content)

	domains, err := LoadLocalFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d: %v", len(domains), domains)
	}
}

func TestLoadLocalFile_EmptyFile(t *testing.T) {
	f := writeTempFile(t, "")
	domains, err := LoadLocalFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}

// ---- LoadRemote ----

func TestLoadRemote_FileNotFound(t *testing.T) {
	_, err := LoadRemote("/nonexistent/path/to/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadRemote_EmptyFile(t *testing.T) {
	f := writeTempFile(t, "")
	domains, err := LoadRemote(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}

func TestLoadRemote_SkipsCommentsAndBlankLines(t *testing.T) {
	f := writeTempFile(t, "# comment\n\n   \n")
	domains, err := LoadRemote(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}

func TestLoadRemote_DownloadsAndParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "# remote blocklist")
		fmt.Fprintln(w, "0.0.0.0 remote-bad.com")
		fmt.Fprintln(w, "remote-also-bad.com")
	}))
	defer srv.Close()

	f := writeTempFile(t, srv.URL+"\n")

	domains, err := LoadRemote(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d: %v", len(domains), domains)
	}
}

func TestLoadRemote_PartialFailure(t *testing.T) {
	// One valid server, one bad URL — should still return domains from the valid one
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "good.com")
	}))
	defer srv.Close()

	f := writeTempFile(t, "http://127.0.0.1:0/nonexistent\n"+srv.URL+"\n")

	domains, err := LoadRemote(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 1 || domains[0] != "good.com" {
		t.Errorf("expected [good.com], got %v", domains)
	}
}

func TestLoadRemote_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	f := writeTempFile(t, srv.URL+"\n")

	// Should not return an error (download failure is logged, not propagated)
	domains, err := LoadRemote(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains on 404, got %d: %v", len(domains), domains)
	}
}

// ---- helpers ----

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "blocklist-*.txt")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return filepath.Clean(f.Name())
}
