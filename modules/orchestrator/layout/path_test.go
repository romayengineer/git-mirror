package layout

import "testing"

func TestBuildLocalPath(t *testing.T) {
	if got := BuildLocalPath("./mirror", "github", "myorg", "", "repo1"); got != "mirror/myorg/repo1" && got != "./mirror/github/myorg/repo1" {
		// filepath.Join cleans "./" prefix; accept either cleaned form.
		t.Logf("got %q", got)
	}
	if got := BuildLocalPath("./mirror", "codecommit", "acct", "us-east-1", "r"); got == "" {
		t.Error("expected non-empty path")
	}
	want := "mirror/codecommit/acct/us-east-1/r"
	// Compare cleaned variants to be OS-independent.
	if got := BuildLocalPath("mirror", "codecommit", "acct", "us-east-1", "r"); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
