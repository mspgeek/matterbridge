package bdiscord

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

// relayedMessageTypes are the Discord message types people write. Everything
// else is a system message (thread created or renamed, pins, joins, boosts).
var relayedMessageTypes = map[discordgo.MessageType]bool{
	discordgo.MessageTypeDefault:            true,
	discordgo.MessageTypeReply:              true,
	discordgo.MessageTypeChatInputCommand:   true,
	discordgo.MessageTypeContextMenuCommand: true,
}

// maxThreadTitle is Discord's limit for channel and thread names.
const maxThreadTitle = 100

var (
	userMentionMarkupRE = regexp.MustCompile(`<@!?([0-9]+)>`)
	roleMentionMarkupRE = regexp.MustCompile(`<@&([0-9]+)>`)
	threadLinkRE        = regexp.MustCompile(`<#([^>\n]+)>`)
	threadTitleMarkdown = strings.NewReplacer("**", "", "__", "", "~~", "", "||", "", "`", "")
)

// threadTitle names a thread started on m: the message text as people see
// it, with mentions, channels and emotes shown by name, on one line and cut
// to Discord's limit.
func (b *Bdiscord) threadTitle(m *discordgo.Message) string {
	text := m.Content
	if text == "" {
		for _, attach := range m.Attachments {
			if attach != nil && attach.Filename != "" {
				text = attach.Filename
				break
			}
		}
	}
	text = userMentionMarkupRE.ReplaceAllStringFunc(text, func(match string) string {
		id := userMentionMarkupRE.FindStringSubmatch(match)[1]
		for _, user := range m.Mentions {
			if user != nil && user.ID == id {
				return "@" + b.displayName(user, b.guildID)
			}
		}
		if member := b.memberByID(b.guildID, id); member != nil && member.User != nil {
			return "@" + b.displayName(member.User, b.guildID)
		}
		return "@unknown-user"
	})
	text = roleMentionMarkupRE.ReplaceAllStringFunc(text, func(match string) string {
		id := roleMentionMarkupRE.FindStringSubmatch(match)[1]
		if role, err := b.c.State.Role(b.guildID, id); err == nil && role != nil {
			return "@" + role.Name
		}
		return "@unknown-role"
	})
	text = b.replaceChannelMentions(text)
	text = threadLinkRE.ReplaceAllString(text, "#$1") // replaceChannelMentions shows threads as "<#name>"
	text = replaceEmotes(text)
	text = threadTitleMarkdown.Replace(text)
	return truncateTitle(strings.Join(strings.Fields(text), " "), maxThreadTitle)
}

// displayName is the name Discord shows for user in the guild: server
// nickname, then global display name, then username.
func (b *Bdiscord) displayName(user *discordgo.User, guildID string) string {
	if member := b.memberByID(guildID, user.ID); member != nil && member.Nick != "" {
		return member.Nick
	}
	if user.GlobalName != "" {
		return user.GlobalName
	}
	return user.Username
}

// truncateTitle cuts s to at most n characters, ending in "…" when cut.
func truncateTitle(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:n-1])) + "…"
}
