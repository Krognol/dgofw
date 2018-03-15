package dgofw

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordGuild struct {
	client *DiscordClient
	g      *discordgo.Guild
	Colors map[string]int
	Owner  *DiscordMember
}

func NewDiscordGuild(client *DiscordClient, g *discordgo.Guild) *DiscordGuild {
	result := &DiscordGuild{
		client: client,
		g:      g,
		Colors: make(map[string]int),
		Owner:  nil,
	}
	if owner := client.Cache.GetMember(g.ID, g.OwnerID); owner != nil {
		result.Owner = owner
	} else {
		m, _ := client.ses.GuildMember(g.ID, g.OwnerID)
		result.Owner = NewDiscordMember(client, m)
	}
	return result
}

func (g *DiscordGuild) ID() string {
	return g.g.ID
}

func (g *DiscordGuild) VoiceStates() []*discordgo.VoiceState {
	return g.g.VoiceStates
}

func (g *DiscordGuild) Emojis() []*discordgo.Emoji {
	return g.g.Emojis
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

func (g *DiscordGuild) Client() *DiscordClient {
	return g.client
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
	return time.Unix(ms/1000, 0).UTC().Format("2006-01-02 15:04:05")
}

func (g *DiscordGuild) Edit(params discordgo.GuildParams) {
	g.client.ses.GuildEdit(g.ID(), params)
}

func (g *DiscordGuild) Members() []*DiscordMember {
	iter := g.g.Members
	result := make([]*DiscordMember, len(iter))
	for i, m := range iter {
		if cm := g.client.Cache.GetMember(g.ID(), m.User.ID); cm != nil {
			result[i] = cm
		} else {
			result[i] = NewDiscordMember(g.client, m)
		}
	}
	return result
}

func (g *DiscordGuild) Member(id string) *DiscordMember {
	if mem := g.client.Cache.GetMember(g.ID(), id); mem != nil {
		return mem
	}
	for _, mem := range g.g.Members {
		if mem.User.ID == id {
			return NewDiscordMember(g.client, mem)
		}
	}
	return nil
}

func (g *DiscordGuild) Channels() []*DiscordChannel {
	iter := g.g.Channels
	result := make([]*DiscordChannel, len(iter))
	for i, c := range iter {
		if cc := g.client.Cache.GetChannel(c.ID); cc != nil {
			result[i] = cc
		} else {
			result[i] = NewDiscordChannel(g.client, c)
		}
	}
	return result
}
