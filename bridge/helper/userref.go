package helper

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
)

// A user reference marks a real @mention from the source platform. It carries
// every name of the mentioned user (the first is the display name) so the
// destination can link the right user without guessing where a name ends in
// plain text. The names are base64url-encoded JSON, so no character in a name
// can break the marker: "<@user:WyJNZW5keSBHcmVlbiJd>".
var userRefRE = regexp.MustCompile(`<@user:([A-Za-z0-9_-]+)>`)

// UserRef returns the reference for a user. The first non-empty name is shown
// when the destination cannot link the user.
func UserRef(names ...string) string {
	var clean []string
	for _, name := range names {
		if name != "" {
			clean = append(clean, name)
		}
	}
	if len(clean) == 0 {
		return ""
	}
	data, _ := json.Marshal(clean)
	return "<@user:" + base64.RawURLEncoding.EncodeToString(data) + ">"
}

// ReplaceUserRefs resolves every user reference in text. resolve returns the
// destination's mention for the user with these names, or false; unresolved
// references become "@" + display name.
func ReplaceUserRefs(text string, resolve func(names []string) (string, bool)) string {
	return userRefRE.ReplaceAllStringFunc(text, func(match string) string {
		names := decodeUserRef(match)
		if len(names) == 0 {
			return match
		}
		if resolve != nil {
			if mention, ok := resolve(names); ok {
				return mention
			}
		}
		return "@" + names[0]
	})
}

// HasUserRefs reports whether text contains user references.
func HasUserRefs(text string) bool {
	return userRefRE.MatchString(text)
}

func decodeUserRef(match string) []string {
	sub := userRefRE.FindStringSubmatch(match)
	if sub == nil {
		return nil
	}
	data, err := base64.RawURLEncoding.DecodeString(sub[1])
	if err != nil {
		return nil
	}
	var names []string
	if json.Unmarshal(data, &names) != nil {
		return nil
	}
	return names
}

// nameDecorations separate a name from a title or company added to it, as in
// "Kelvin Tegelaar - MVP/PoshSpice" or "Mendy Green [MSPGeek]".
var nameDecorations = []string{" - ", " – ", " — ", " | ", " [", " ("}

// stripDecoration returns name without a decoration, or "" when it has none.
func stripDecoration(name string) string {
	cut := -1
	for _, sep := range nameDecorations {
		if i := strings.Index(name, sep); i > 0 && (cut < 0 || i < cut) {
			cut = i
		}
	}
	if cut < 0 {
		return ""
	}
	return strings.TrimSpace(name[:cut])
}

// wordCount returns the number of words in name that contain a letter or digit.
func wordCount(name string) int {
	n := 0
	for _, word := range strings.Fields(name) {
		if NormalizeName(word) != "" {
			n++
		}
	}
	return n
}

// ResolveUserByNames returns the user ID matched by the most specific of names.
// A name matches only when the whole name matches a user in idx; the number of
// words in the name ranks the matches, so "Kelvin Tegelaar" (two words) wins
// over a username "kelvin" that belongs to someone else. Each name is also
// tried without a decoration such as " - Title" or " [Company]". When the best
// matches point to different users, nothing is linked.
func ResolveUserByNames(idx NameIndex, names []string) (string, bool) {
	var candidates []string
	for _, name := range names {
		candidates = append(candidates, name)
		if stripped := stripDecoration(name); stripped != "" {
			candidates = append(candidates, stripped)
		}
	}
	best, bestScore, tied := "", 0, false
	for _, name := range candidates {
		id, ok := idx.Lookup(name)
		if !ok {
			continue
		}
		switch score := wordCount(name); {
		case score > bestScore:
			best, bestScore, tied = id, score, false
		case score == bestScore && id != best:
			tied = true
		}
	}
	if best == "" || tied {
		return "", false
	}
	return best, true
}
