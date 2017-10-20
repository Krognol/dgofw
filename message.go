package dgofw

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type DiscordMessage struct {
	sync.RWMutex
	keys     []string
	vals     []string
	m        *discordgo.Message
	s        *discordgo.Session
	Author   *DiscordUser
	Mentions []*DiscordUser
}

func (m *DiscordMessage) ID() string {
	return m.m.ID
}

func (m *DiscordMessage) ChannelID() string {
	return m.m.ChannelID
}

func (m *DiscordMessage) GuildID() string {
	return m.Channel().GuildID()
}

func (m *DiscordMessage) Timestamp() string {
	result, _ := m.m.Timestamp.Parse()
	return result.String()
}

func (m *DiscordMessage) Content() string {
	return m.m.Content
}

func (m *DiscordMessage) Session() *discordgo.Session {
	return m.s
}

func NewDiscordMessage(s *discordgo.Session, m *discordgo.Message) *DiscordMessage {
	result := &DiscordMessage{
		keys:     make([]string, 0),
		vals:     make([]string, 0),
		m:        m,
		s:        s,
		Author:   NewDiscordUser(s, m.Author),
		Mentions: make([]*DiscordUser, 0),
	}

	if len(m.Mentions) > 0 {
		result.Mentions = make([]*DiscordUser, len(m.Mentions))
		for i, u := range m.Mentions {
			result.Mentions[i] = NewDiscordUser(s, u)
		}
	}
	return result
}

func (m *DiscordMessage) IsMod() bool {
	perms, _ := m.s.State.UserChannelPermissions(m.Author.ID(), m.ChannelID())

	return ((perms & discordgo.PermissionAll) == discordgo.PermissionAll) ||
		((perms & discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator) ||
		((perms & discordgo.PermissionManageServer) == discordgo.PermissionManageServer) ||
		((perms & discordgo.PermissionAllChannel) == discordgo.PermissionAllChannel)
}

func (m *DiscordMessage) AddKey(key string) {
	m.RLock()
	m.keys = append(m.keys, key)
	m.RUnlock()
}

func (m *DiscordMessage) AddVal(val string) {
	m.RLock()
	m.vals = append(m.vals, val)
	m.RUnlock()
}

func (m *DiscordMessage) Reply(msg string) *DiscordMessage {
	m2, err := m.s.ChannelMessageSend(m.ChannelID(), msg)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return NewDiscordMessage(m.s, m2)
}

func (m *DiscordMessage) ReplyComplex(m2 *discordgo.MessageSend) *DiscordMessage {
	m3, err := m.s.ChannelMessageSendComplex(m.ChannelID(), m2)
	if err != nil {
		return nil
	}
	return NewDiscordMessage(m.s, m3)
}

func (m *DiscordMessage) ReplyEmbed(embed *discordgo.MessageEmbed) *DiscordMessage {
	m2, err := m.s.ChannelMessageSendEmbed(m.ChannelID(), embed)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return NewDiscordMessage(m.s, m2)
}

func (m *DiscordMessage) ReplyFile(name string, file io.Reader) *DiscordMessage {
	m2, err := m.s.ChannelFileSend(m.ChannelID(), name, file)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return NewDiscordMessage(m.s, m2)
}

func (m *DiscordMessage) ReplyFileWithMessage(msg, name string, file io.Reader) *DiscordMessage {
	m2, err := m.s.ChannelFileSendWithMessage(m.ChannelID(), msg, name, file)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return NewDiscordMessage(m.s, m2)
}

func (m *DiscordMessage) Arg(key string) string {
	for i, s := range m.keys {
		if s == key {
			return m.vals[i]
		}
	}
	return ""
}

func (m *DiscordMessage) Edit(msg string) {
	m.s.ChannelMessageEdit(m.ChannelID(), m.ID(), msg)
}

func (m *DiscordMessage) EditEmbed(embed *discordgo.MessageEmbed) {
	s := new(string)
	*s = ""
	_, err := m.s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content: s,
		Embed:   embed,
		ID:      m.ID(),
		Channel: m.ChannelID(),
	})

	if err != nil {
		fmt.Println(err)
	}
}

func (m *DiscordMessage) Delete() {
	m.s.ChannelMessageDelete(m.ChannelID(), m.ID())
}

func (m *DiscordMessage) DeleteManyIDs(ids ...string) {
	err := m.s.ChannelMessagesBulkDelete(m.ChannelID(), ids)
	if err != nil {
		fmt.Println(err)
	}
}

func (m *DiscordMessage) DeleteMany(msgs ...*DiscordMessage) {
	result := make([]string, len(msgs))
	for i, msg := range msgs {
		result[i] = msg.ID()
	}
	m.DeleteManyIDs(result...)
}

func (m *DiscordMessage) Pairs(keys, vals []string) {
	m.keys = make([]string, len(keys))
	m.vals = make([]string, len(vals))
	for i, key := range keys {
		m.keys[i] = strings.Trim(key, "{}")
	}

	for i := 0; i < len(keys); i++ {
		if i == len(keys)-1 {
			if i <= len(vals)-1 {
				m.vals[i] = strings.Join(vals[i:], " ")
			} else {
				m.vals = append(m.vals, "")
			}
		} else {
			if i <= len(vals)-1 {
				m.vals[i] = vals[i]
			} else {
				m.vals = append(m.vals, "")
			}
		}
	}
}

func (m *DiscordMessage) Channel() *DiscordChannel {
	if c := Cache.GetChannel(m.ChannelID()); c != nil {
		return c
	}
	c, err := m.Session().Channel(m.ChannelID())
	if err != nil {
		return nil
	}
	return NewDiscordChannel(m.Session(), c)
}

func (m *DiscordMessage) PrintPairs() {
	for i := 0; i < len(m.keys); i++ {
		fmt.Printf("[%d]:\t%s:%s\n", i, m.keys[i], m.vals[i])
	}
}

func (m *DiscordMessage) Guild() *DiscordGuild {
	if ch := m.Channel(); ch != nil {
		if gm := Cache.GetGuild(ch.GuildID()); gm != nil {
			return gm
		}
		g, _ := m.s.Guild(ch.GuildID())
		return NewDiscordGuild(m.Session(), g)
	}
	return nil
}

// WaitForMessage intercepts messages until ``timeout`` is reached, or ``cb`` returns ``true``.
func (m *DiscordMessage) WaitForMessage(timeout int, cb func(*DiscordMessage) bool, onTimeout func()) {
	Cache.client.waitForMessage(timeout, cb, onTimeout)
}

func (m *DiscordMessage) WaitForever(cb func(*DiscordMessage)) (done chan bool) {
	return Cache.client.waitForever(cb)
}

func (m *DiscordMessage) React(emoji string) {
	m.s.MessageReactionAdd(m.ChannelID(), m.ID(), emoji)
}

func (m *DiscordMessage) RemoveReaction(emoji string) {
	m.s.MessageReactionRemove(m.ChannelID(), m.ID(), emoji, m.Author.ID())
}

func (m *DiscordMessage) HasMention() bool {
	return len(m.m.Mentions) > 0
}

func (m *DiscordMessage) MentionsEveryone() bool {
	return m.m.MentionEveryone
}
