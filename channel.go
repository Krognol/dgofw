package dgofw

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type DiscordChannel struct {
	c      *discordgo.Channel
	client *DiscordClient
	Guild  *DiscordGuild
}

func NewDiscordChannel(client *DiscordClient, c *discordgo.Channel) *DiscordChannel {
	return &DiscordChannel{
		c:      c,
		client: client,
		Guild:  client.Cache.GetGuild(c.GuildID),
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

func (c *DiscordChannel) Send(msg string) *DiscordMessage {
	m, err := c.client.ses.ChannelMessageSend(c.ID(), msg)
	if err != nil {
		return nil
	}

	return NewDiscordMessage(c.client, m)
}

func (c *DiscordChannel) Messages(count int, before, after, around string) []*DiscordMessage {
	msgs, err := c.client.ses.ChannelMessages(c.ID(), count, before, after, around)
	if err != nil {
		return nil
	}
	result := make([]*DiscordMessage, len(msgs))
	for i, m := range msgs {
		result[i] = NewDiscordMessage(c.client, m)
	}
	return result
}

func (c *DiscordChannel) DeleteMessages(count int) {
	msgs, err := c.client.ses.ChannelMessages(c.ID(), count, "", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	result := make([]string, len(msgs))
	for i, msg := range msgs {
		result[i] = msg.ID
	}

	if err = c.client.ses.ChannelMessagesBulkDelete(c.ID(), result); err != nil {
		fmt.Println(err)
	}
}

func (c *DiscordChannel) Message(id string) *DiscordMessage {
	m, err := c.client.ses.ChannelMessage(c.ID(), id)
	if err != nil {
		return nil
	}
	return NewDiscordMessage(c.client, m)
}

func (c *DiscordChannel) PinMessage(id string) {
	c.client.ses.ChannelMessagePin(c.ID(), id)
}
