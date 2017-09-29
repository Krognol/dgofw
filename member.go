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
	Guild *DiscordGuild
	Roles []*discordgo.Role
}

func NewDiscordMember(s *discordgo.Session, m *discordgo.Member) *DiscordMember {
	if mem := Cache.GetMember(m.User.ID); mem != nil {
		return mem
	}
	result := &DiscordMember{
		s:     s,
		m:     m,
		Roles: make([]*discordgo.Role, 0),
	}
	if g := Cache.GetGuild(m.GuildID); g != nil {
		result.Guild = g
	} else {
		gg, _ := s.Guild(m.GuildID)
		result.Guild = NewDiscordGuild(s, gg)
	}

	result.User = NewDiscordUser(s, m.User)
	for _, role := range result.Guild.Roles() {
		for _, id := range m.Roles {
			if role.ID == id {
				result.Roles = append(result.Roles, role)
			}
		}
	}
	return result
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
	if m.Guild != nil {
		if c, ok := m.Guild.Colors[m.User.ID()]; ok {
			return c
		}
	}
	c := m.s.State.UserColor(m.User.ID(), m.Guild.Channels()[0].ID())
	m.Guild.Colors[m.User.ID()] = c
	return c
}
