package gitsync

import "strings"

// codecommitIdentity reduces a CodeCommit URL to its (region, repository name)
// identity, regardless of the transport used to reach it. It understands:
//
//   - git-remote-codecommit (GRC) helper URLs:
//     codecommit::<region>://<profile>@<name>
//     codecommit::<region>://<name>
//     codecommit://<profile>@<name>
//     codecommit://<name>
//   - SSH/HTTPS clone URLs:
//     ssh://[<user>@]git-codecommit.<region>.amazonaws.com/v1/repos/<name>
//     https://git-codecommit.<region>.amazonaws.com/v1/repos/<name>
//
// The SSH user segment (an SSH key ID) and the GRC named profile are ignored:
// they identify *how* to authenticate, not *which* repository is addressed.
//
// The returned region is lowercased (AWS regions are case-insensitive); the
// name is returned verbatim (CodeCommit repository names are case-sensitive).
// ok is false when the URL is not a recognizable CodeCommit URL.
func codecommitIdentity(raw string) (region, name string, ok bool) {
	u := normalizeURL(raw)
	if u == "" {
		return "", "", false
	}

	if r, n, matched := codecommitRemoteHelperIdentity(u); matched {
		return strings.ToLower(r), n, true
	}
	if r, n, matched := codecommitHostIdentity(u); matched {
		return strings.ToLower(r), n, true
	}

	return "", "", false
}

// codecommitRemoteHelperIdentity parses a GRC URL (codecommit://…).
func codecommitRemoteHelperIdentity(u string) (region, name string, ok bool) {
	if !strings.HasPrefix(u, "codecommit:") {
		return "", "", false
	}

	rest := strings.TrimPrefix(u, "codecommit:")
	switch {
	case strings.HasPrefix(rest, ":"):
		// codecommit::<region>://<profile>@<name>
		rest = strings.TrimPrefix(rest, ":")
		r, after, found := strings.Cut(rest, "://")
		if !found {
			return "", "", false
		}
		region = r
		rest = after
	case strings.HasPrefix(rest, "//"):
		// codecommit://<profile>@<name> (no explicit region)
		rest = strings.TrimPrefix(rest, "//")
	default:
		return "", "", false
	}

	// Drop the optional "<profile>@" userinfo segment.
	rest = stripUserInfo(rest)
	if rest == "" || strings.ContainsAny(rest, ":/") {
		return "", "", false
	}

	return region, rest, true
}

// codecommitHostIdentity parses an ssh:// or https:// CodeCommit clone URL.
func codecommitHostIdentity(u string) (region, name string, ok bool) {
	const pathPrefix = "v1/repos/"

	var rest string
	for _, scheme := range []string{"ssh://", "https://", "http://"} {
		if strings.HasPrefix(u, scheme) {
			rest = strings.TrimPrefix(u, scheme)
			break
		}
	}
	if rest == "" {
		return "", "", false
	}

	rest = stripUserInfo(rest)

	host, path, found := strings.Cut(rest, "/")
	if !found {
		return "", "", false
	}
	if !strings.HasPrefix(host, "git-codecommit.") || !strings.HasSuffix(host, ".amazonaws.com") {
		return "", "", false
	}
	region = strings.TrimSuffix(strings.TrimPrefix(host, "git-codecommit."), ".amazonaws.com")
	if region == "" || !strings.HasPrefix(path, pathPrefix) {
		return "", "", false
	}
	name = strings.TrimPrefix(path, pathPrefix)
	if name == "" {
		return "", "", false
	}

	return region, name, true
}

// stripUserInfo removes a leading "<userinfo>@" segment, if present.
func stripUserInfo(s string) string {
	if at := strings.LastIndex(s, "@"); at != -1 {
		return s[at+1:]
	}
	return s
}
