package dgofw

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type (
	OneOf struct {
		msg   func(*DiscordMessage)
		ch    func(*DiscordChannel)
		mem   func(*DiscordMember)
		guild func(*DiscordGuild)
		usr   func(*DiscordUser)
		rdy   func(*discordgo.Ready)
	}
	Handler struct {
		once    bool
		pattern string
		cb      *OneOf
		typ     DiscordEvent
	}
)

func (c *DiscordClient) delHandler(i int) {
	c.RLock()
	if i > 1 {
		c.handlers = append(c.handlers[:i], c.handlers[i+1:]...)
	} else {
		c.handlers = c.handlers[1:]
	}
	c.RUnlock()

}

func (c *DiscordClient) handleMessageC(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	for i, handler := range c.handlers {
		if handler.typ == evMessage {
			msg := NewDiscordMessage(s, m.Message)
			if len(c.interceptors) > 0 {
				for _, iter := range c.interceptors {
					iter <- msg
				}
			}

			vals := strings.Fields(m.Content)
			if len(vals) > 0 || len(m.Content) >= len(handler.pattern) {
				if m.Content == handler.pattern || strings.HasPrefix(handler.pattern, m.Content[0:len(vals[0])]) {
					keys := strings.Fields(handler.pattern)

					if len(keys) > 0 || len(vals) > 0 {
						msg.Pairs(keys, vals)
					}

					go handler.cb.msg(msg)
					if handler.once {
						c.delHandler(i)
					}
				}
			}
		}
	}
}

func (c *DiscordClient) handleMessageE(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author == nil {
		return
	}
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	c.handleMessageC(s, (*discordgo.MessageCreate)(m))
}

// OnMessage handles a ``MESSAGE_*`` event.
// Does not handle ``MESSAGE_DELETE``
func (c *DiscordClient) OnMessage(pattern string, once bool, cb func(*DiscordMessage)) {
	c.handlers = append(c.handlers, &Handler{
		cb:      &OneOf{msg: cb},
		once:    once,
		pattern: pattern,
		typ:     evMessage,
	})
}

// OnReady handles a Discord ``READY`` event
func (c *DiscordClient) OnReady(once bool, cb func(*discordgo.Ready)) {
	c.handlers = append(c.handlers, &Handler{
		cb:   &OneOf{rdy: cb},
		once: once,
		typ:  evReady,
	})
}

func (c *DiscordClient) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	for i, h := range c.handlers {
		if h.typ == evReady {
			go h.cb.rdy(r)
			if h.once {
				c.delHandler(i)
			}
		}
	}
}

func (c *DiscordClient) handleGuild(s *discordgo.Session, g *discordgo.Guild) {
	for i, h := range c.handlers {
		if h.typ == evGuild {
			go h.cb.guild(NewDiscordGuild(s, g))
			if h.once {
				c.delHandler(i)
			}
		}
	}
}

func (c *DiscordClient) handleGuildE(s *discordgo.Session, g *discordgo.GuildUpdate) {
	Cache.UpdateGuild(s, g.Guild)
	c.handleGuild(s, g.Guild)
}

func (c *DiscordClient) handleGuildD(s *discordgo.Session, g *discordgo.GuildUpdate) {
	Cache.DeleteGuild(g.ID)
}

func (c *DiscordClient) handleChannel(s *discordgo.Session, ch *discordgo.Channel) {
	for i, h := range c.handlers {
		if h.typ == evChannel {
			go h.cb.ch(NewDiscordChannel(s, ch))
			if h.once {
				c.delHandler(i)
			}
		}
	}
}

func (c *DiscordClient) handleChannelC(s *discordgo.Session, ch *discordgo.ChannelCreate) {
	Cache.UpdateChannel(s, ch.Channel)
	c.handleChannel(s, ch.Channel)
}

func (c *DiscordClient) handleChannelE(s *discordgo.Session, ch *discordgo.ChannelUpdate) {
	Cache.UpdateChannel(s, ch.Channel)
	c.handleChannel(s, ch.Channel)
}

func (c *DiscordClient) handleChannelD(s *discordgo.Session, ch *discordgo.ChannelDelete) {
	Cache.DeleteChannel(ch.ID)
	c.handleChannel(s, ch.Channel)
}

func (c *DiscordClient) handleMember(s *discordgo.Session, m *discordgo.Member) {
	for i, h := range c.handlers {
		if h.typ == evMember {
			go h.cb.mem(NewDiscordMember(s, m))
			if h.once {
				c.delHandler(i)
			}
		}
	}
}

func (c *DiscordClient) handleMemberA(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	Cache.UpdateMember(s, m.Member)
	c.handleMember(s, m.Member)
}

func (c *DiscordClient) handleMemberE(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	Cache.UpdateMember(s, m.Member)
	c.handleMember(s, m.Member)
}

func (c *DiscordClient) handleMemberD(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	Cache.DeleteMember(m.User.ID)
	c.handleMember(s, m.Member)
}

// WithMemberChunk handles a ``GuildMembersChunk`` event.
func (c *DiscordClient) WithMemberChunk(cb func(*discordgo.GuildMembersChunk)) {
	chunkCb := func(_ *discordgo.Session, chunk *discordgo.GuildMembersChunk) {
		cb(chunk)
	}
	c.ses.AddHandler(chunkCb)
}
