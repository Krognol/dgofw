package dgofw

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordMember struct {
	client *DiscordClient
	m      *discordgo.Member
	User   *DiscordUser
	Roles  []*discordgo.Role
}

func NewDiscordMember(client *DiscordClient, m *discordgo.Member) *DiscordMember {
	result := &DiscordMember{
		client: client,
		m:      m,
		Roles:  make([]*discordgo.Role, 0),
	}

	result.User = NewDiscordUser(client, m.User)
	return result
}

func (m *DiscordMember) Guild() *DiscordGuild {
	if res := m.client.Cache.GetGuild(m.m.GuildID); res != nil {
		return res
	}
	g, err := m.client.Cache.client.ses.Guild(m.m.GuildID)
	if err != nil {
		return nil
	}
	return NewDiscordGuild(m.client, g)
}

func (m *DiscordMember) GuildID() string {
	return m.m.GuildID
}

func (m *DiscordMember) Nickname() string {
	return m.m.Nick
}

func (m *DiscordMember) JoinedAt() string {
	t, err := time.Parse(time.RFC3339Nano, m.m.JoinedAt)
	if err != nil {
		fmt.Println(err)
		return "<nil>"
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

func (m *DiscordMember) Color() int {
	g := m.Guild()
	if g != nil {
		if c, ok := g.Colors[m.User.ID()]; ok {
			return c
		}
	}
	c := m.client.ses.State.UserColor(m.User.ID(), g.Channels()[0].ID())
	g.Colors[m.User.ID()] = c
	return c
}

func (m *DiscordMember) Ban(days int) error {
	return m.client.ses.GuildBanCreate(m.GuildID(), m.User.ID(), days)
}
