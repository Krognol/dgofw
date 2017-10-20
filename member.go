package dgofw

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordMember struct {
	s     *discordgo.Session
	m     *discordgo.Member
	User  *DiscordUser
	Roles []*discordgo.Role
}

func NewDiscordMember(s *discordgo.Session, m *discordgo.Member) *DiscordMember {
	result := &DiscordMember{
		s:     s,
		m:     m,
		Roles: make([]*discordgo.Role, 0),
	}

	result.User = NewDiscordUser(s, m.User)
	return result
}

func (m *DiscordMember) Guild() *DiscordGuild {
	if res := Cache.GetGuild(m.m.GuildID); res != nil {
		return res
	}
	g, err := Cache.client.ses.Guild(m.m.GuildID)
	if err != nil {
		return nil
	}
	return NewDiscordGuild(m.s, g)
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
	return t.Format("2006-01-02 15:04:05")
}

func (m *DiscordMember) Color() int {
	g := m.Guild()
	if g != nil {
		if c, ok := g.Colors[m.User.ID()]; ok {
			return c
		}
	}
	c := m.s.State.UserColor(m.User.ID(), g.Channels()[0].ID())
	g.Colors[m.User.ID()] = c
	return c
}

func (m *DiscordMember) Ban(days int) error {
	return m.s.GuildBanCreate(m.GuildID(), m.User.ID(), days)
}
