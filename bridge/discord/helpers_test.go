package bdiscord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestReplaceUserMentions(t *testing.T) {
	b := &Bdiscord{
		userMemberMap: map[string]*discordgo.Member{},
		nickMemberMap: map[string]*discordgo.Member{},
	}
	b.addMemberLocked(&discordgo.Member{User: &discordgo.User{ID: "1", Username: "jsmith", GlobalName: "John Smith"}})
	b.addMemberLocked(&discordgo.Member{User: &discordgo.User{ID: "2", Username: "kdoe"}, Nick: "Kim Doe"})

	// Slack sends display names with spaces removed.
	assert.Equal(t, "hey <@1>, see <@2>", b.replaceUserMentions("hey @JohnSmith, see @KimDoe"))
	assert.Equal(t, "<@1> and <@2>", b.replaceUserMentions("@jsmith and @kim doe"))
	assert.Equal(t, "@nobody", b.replaceUserMentions("@nobody"))

	// The index follows member updates.
	b.addMemberLocked(&discordgo.Member{User: &discordgo.User{ID: "3", Username: "newbie"}})
	assert.Equal(t, "<@3>", b.replaceUserMentions("@newbie"))
}
