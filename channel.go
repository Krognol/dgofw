package dgofw

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordChannel struct {
	c     *discordgo.Channel
	s     *discordgo.Session
	Guild *DiscordGuild
}

func NewDiscordChannel(s *discordgo.Session, c *discordgo.Channel) *DiscordChannel {
	return &DiscordChannel{
		c:     c,
		s:     s,
		Guild: Cache.GetGuild(c.GuildID),
	}
}

func (c *DiscordChannel) ID() string {
	return c.c.ID
}

func (c *DiscordChannel) NSFW() bool {
	return c.c.NSFW
}

func (c *DiscordChannel) Name() string {
	return c.c.Name
}

func (c *DiscordChannel) Position() int {
	return c.c.Position
}

func (c *DiscordChannel) Topic() string {
	return c.c.Topic
}

func (c *DiscordChannel) GuildID() string {
	return c.c.GuildID
}

func (c *DiscordChannel) Type() discordgo.ChannelType {
	return c.c.Type
}
