package bdiscord

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/mspgeek-community/matterbridge/bridge"
	"github.com/mspgeek-community/matterbridge/bridge/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemMessagesAreNotRelayed(t *testing.T) {
	for _, typ := range []discordgo.MessageType{
		discordgo.MessageTypeThreadCreated,
		discordgo.MessageTypeChannelNameChange,
		discordgo.MessageTypeChannelPinnedMessage,
		discordgo.MessageTypeGuildMemberJoin,
		discordgo.MessageTypeThreadStarterMessage,
	} {
		assert.False(t, relayedMessageTypes[typ], "type %d", typ)
	}
	for _, typ := range []discordgo.MessageType{discordgo.MessageTypeDefault, discordgo.MessageTypeReply} {
		assert.True(t, relayedMessageTypes[typ], "type %d", typ)
	}

	// relayMessage drops a "thread created" message before doing anything else.
	b := &Bdiscord{Config: &bridge.Config{Remote: make(chan config.Message, 1)}}
	b.relayMessage(&discordgo.MessageCreate{Message: &discordgo.Message{
		Type:    discordgo.MessageTypeThreadCreated,
		Content: "Thread name",
		Author:  &discordgo.User{ID: "1", Username: "someone"},
	}}, false)
	assert.Empty(t, b.Remote)
}

func TestThreadTitle(t *testing.T) {
	session, err := discordgo.New("Bot test")
	require.NoError(t, err)
	require.NoError(t, session.State.GuildAdd(&discordgo.Guild{ID: "g", Roles: []*discordgo.Role{{ID: "77", Name: "MSPGeek Admins"}}}))
	b := &Bdiscord{
		c:             session,
		guildID:       "g",
		userMemberMap: map[string]*discordgo.Member{},
		nickMemberMap: map[string]*discordgo.Member{},
	}
	b.addMemberLocked(&discordgo.Member{User: &discordgo.User{ID: "111", Username: "tegek", GlobalName: "KelvinTegelaar"}})
	b.addMemberLocked(&discordgo.Member{User: &discordgo.User{ID: "222", Username: "mendyg", GlobalName: "Mendy Green"}, Nick: "Mendy Green [MSPGeek]"})

	m := &discordgo.Message{
		Content:  "<@111> <@!222> did a **PR** for <@&77> :wave: <:hudu:123456>\nsee <#my-thread>",
		Mentions: []*discordgo.User{{ID: "111", Username: "tegek", GlobalName: "KelvinTegelaar"}},
	}
	assert.Equal(t, "@KelvinTegelaar @Mendy Green [MSPGeek] did a PR for @MSPGeek Admins :wave: :hudu: see #my-thread", b.threadTitle(m))

	// Cut to Discord's 100-character limit on a character boundary.
	long := strings.Repeat("é", 150)
	title := b.threadTitle(&discordgo.Message{Content: long})
	assert.Equal(t, 100, len([]rune(title)))
	assert.True(t, strings.HasSuffix(title, "…"))

	// No text: the first attachment's name.
	assert.Equal(t, "screenshot.png", b.threadTitle(&discordgo.Message{Attachments: []*discordgo.MessageAttachment{{Filename: "screenshot.png"}}}))
	assert.Equal(t, "", b.threadTitle(&discordgo.Message{}))
}
