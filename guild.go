package dgofw

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordGuild struct {
	s      *discordgo.Session
	g      *discordgo.Guild
	Colors map[string]int
	Emojis []*discordgo.Emoji
	Owner  *DiscordMember
}

func NewDiscordGuild(s *discordgo.Session, g *discordgo.Guild) *DiscordGuild {
	result := &DiscordGuild{
		s:      s,
		g:      g,
		Colors: make(map[string]int),
		Emojis: g.Emojis,
	}
	if owner := Cache.GetMember(g.OwnerID); owner != nil {
		result.Owner = owner
	} else {
		m, _ := s.GuildMember(g.ID, g.OwnerID)
		result.Owner = NewDiscordMember(s, m)
	}
	return result
}

func (g *DiscordGuild) ID() string {
	return g.g.ID
}

func (g *DiscordGuild) Icon() string {
	return discordgo.EndpointGuildIcon(g.ID(), g.g.Icon)
}

func (g *DiscordGuild) Name() string {
	return g.g.Name
}

func (g *DiscordGuild) Roles() []*discordgo.Role {
	return g.g.Roles
}

func (g *DiscordGuild) Session() *discordgo.Session {
	return g.s
}

func (g *DiscordGuild) MemberCount() int {
	return g.g.MemberCount
}

func (g *DiscordGuild) OwnerID() string {
	return g.g.OwnerID
}

func (g *DiscordGuild) Region() string {
	return g.g.Region
}

func (g *DiscordGuild) CreatedAt() string {
	id, _ := strconv.ParseInt(g.ID(), 10, 64)
	ms := (id >> 22) + 1420070400000
	return time.Unix(ms/1000, 0).Format("2006-01-02 15:04:05")
}

func (g *DiscordGuild) Edit(params discordgo.GuildParams) {
	g.s.GuildEdit(g.ID(), params)
}

func (g *DiscordGuild) Members() []*DiscordMember {
	iter := g.g.Members
	result := make([]*DiscordMember, len(iter))
	for i, m := range iter {
		if cm := Cache.GetMember(m.User.ID); cm != nil {
			result[i] = cm
		} else {
			result[i] = NewDiscordMember(g.Session(), m)
		}
	}
	return result
}

func (g *DiscordGuild) Member(id string) *DiscordMember {
	if mem := Cache.GetMember(id); mem != nil {
		return mem
	}
	for _, mem := range g.g.Members {
		if mem.User.ID == id {
			return NewDiscordMember(g.s, mem)
		}
	}
	return nil
}

func (g *DiscordGuild) Channels() []*DiscordChannel {
	iter := g.g.Channels
	result := make([]*DiscordChannel, len(iter))
	for i, c := range iter {
		if cc := Cache.GetChannel(c.ID); cc != nil {
			result[i] = cc
		} else {
			result[i] = NewDiscordChannel(g.Session(), c)
		}
	}
	return result
}
