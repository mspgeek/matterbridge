package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRefRoundTrip(t *testing.T) {
	ref := UserRef("Mendy Green [MSPGeek]", "", "mendy.green", "J.R. <@weird|name>\nx")
	assert.True(t, HasUserRefs("hi "+ref))
	assert.Equal(t, "", UserRef("", ""))

	var got []string
	out := ReplaceUserRefs("hi "+ref+"!", func(names []string) (string, bool) {
		got = names
		return "<@42>", true
	})
	assert.Equal(t, "hi <@42>!", out)
	assert.Equal(t, []string{"Mendy Green [MSPGeek]", "mendy.green", "J.R. <@weird|name>\nx"}, got)

	// Unresolved and non-linking destinations show the display name.
	assert.Equal(t, "hi @Mendy Green [MSPGeek]!", ReplaceUserRefs("hi "+ref+"!", nil))
	assert.Equal(t, "hi @Mendy Green [MSPGeek]!", ReplaceUserRefs("hi "+ref+"!",
		func([]string) (string, bool) { return "", false }))

	// Plain text and other markup are untouched.
	assert.Equal(t, "@John <@123> <#channel:x>", ReplaceUserRefs("@John <@123> <#channel:x>", nil))
}

func TestResolveUserByNames(t *testing.T) {
	idx := NameIndex{}
	idx.Add("D1", "mendygreen", "Mendy Green")
	idx.Add("D2", "jsmith")

	id, ok := ResolveUserByNames(idx, []string{"Mendy Green [MSPGeek]", "mendy.green", "nomatch"})
	assert.True(t, ok)
	assert.Equal(t, "D1", id)

	_, ok = ResolveUserByNames(idx, []string{"nobody"})
	assert.False(t, ok)

	// A name only matches whole: "John Smith" does not match a user "John".
	idx.Add("D3", "John")
	_, ok = ResolveUserByNames(idx, []string{"John Smith"})
	assert.False(t, ok)
}

func TestResolveUserByNamesPrefersMostWords(t *testing.T) {
	// Discord: "KelvinTegelaar" (tegek) and "Kelvin" (paraphen).
	idx := NameIndex{}
	idx.Add("tegek", "tegek", "KelvinTegelaar")
	idx.Add("paraphen", "paraphen", "Kelvin")

	// Slack: display name, real name, profile real name, username.
	names := []string{"Kelvin Tegelaar - MVP/PoshSpice", "Kelvin Tegelaar", "Kelvin Tegelaar", "kelvin"}
	id, ok := ResolveUserByNames(idx, names)
	assert.True(t, ok)
	assert.Equal(t, "tegek", id)

	// Order does not matter: the two-word name still wins.
	id, ok = ResolveUserByNames(idx, []string{"kelvin", "Kelvin Tegelaar"})
	assert.True(t, ok)
	assert.Equal(t, "tegek", id)

	// Only the one-word name matches: it links.
	id, ok = ResolveUserByNames(idx, []string{"Someone Else", "kelvin"})
	assert.True(t, ok)
	assert.Equal(t, "paraphen", id)
}

func TestResolveUserByNamesTie(t *testing.T) {
	idx := NameIndex{}
	idx.Add("A", "Bob Jones")
	idx.Add("B", "Robert Smith")
	_, ok := ResolveUserByNames(idx, []string{"Bob Jones", "Robert Smith"})
	assert.False(t, ok, "equally specific names for different users must not link")

	// The same user matched twice is not a tie.
	idx.Add("C", "Ann Lee", "annlee")
	id, ok := ResolveUserByNames(idx, []string{"Ann Lee", "Ann-Lee"})
	assert.True(t, ok)
	assert.Equal(t, "C", id)
}

func TestResolveUserByNamesDecorations(t *testing.T) {
	idx := NameIndex{}
	idx.Add("U", "Jane Doe")
	for _, name := range []string{
		"Jane Doe - CTO", "Jane Doe – CTO", "Jane Doe — CTO", "Jane Doe | Acme",
		"Jane Doe [Acme]", "Jane Doe (she/her)",
	} {
		id, ok := ResolveUserByNames(idx, []string{name})
		assert.True(t, ok, name)
		assert.Equal(t, "U", id, name)
	}

	// The full decorated name still matches a member who uses it.
	idx.Add("V", "Jim Bo - MVP")
	id, ok := ResolveUserByNames(idx, []string{"Jim Bo - MVP"})
	assert.True(t, ok)
	assert.Equal(t, "V", id)

	// Hyphenated names are not decorations.
	assert.Equal(t, "", stripDecoration("Mary-Jane Watson"))
	assert.Equal(t, "", stripDecoration("- leading"))
}

func TestReplaceMentionsSkipsMentionMarkup(t *testing.T) {
	resolve := func(name string) (string, bool) { return "<@X>", true }
	assert.Equal(t, "<@123> <@!456>", ReplaceMentions("<@123> <@!456>", resolve))
}
