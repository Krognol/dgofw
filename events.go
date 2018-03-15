package dgofw

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type (
	MsgHandler struct {
		once    bool
		pattern string
		cb      func(*DiscordMessage)
	}

	DiscordGuildBan struct {
		User  *DiscordUser
		Guild *DiscordGuild
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

	msg := NewDiscordMessage(c, m.Message)
	vals := strings.Fields(m.Content)
	if len(c.interceptors) > 0 {
		for _, iter := range c.interceptors {
			iter <- msg
		}
	}

	for i, handler := range c.handlers {
		keys := strings.Fields(handler.pattern)
		if len(vals) > 0 || len(m.Content) >= len(handler.pattern) {
			if strings.ToLower(m.Content) == strings.ToLower(handler.pattern) ||
				strings.HasPrefix(strings.ToLower(vals[0]), strings.ToLower(keys[0])) {

				if len(keys) > 0 || len(vals) > 0 {
					msg.Pairs(keys, vals)
				}

				go handler.cb(msg)
				if handler.once {
					c.delHandler(i)
				}
			}
		}
	}
}

func (c *DiscordClient) handleMessageE(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author != nil {
		c.handleMessageC(s, (*discordgo.MessageCreate)(m))
	}
}

// OnMessage handles a ``MESSAGE_*`` event.
// Does not handle ``MESSAGE_DELETE``
func (c *DiscordClient) OnMessage(pattern string, once bool, cb func(*DiscordMessage)) {
	c.handlers = append(c.handlers, &MsgHandler{
		cb:      cb,
		once:    once,
		pattern: pattern,
	})
}

// OnMessageDeleted handles a ``MESSAGE_DELETE`` event
//
// ``MESSAGE_DELETE`` is a special case, since only 2 fields are present.
func (c *DiscordClient) OnMessageDeleted(once bool, cb func(*discordgo.MessageDelete)) {
	mdcb := func(_ *discordgo.Session, m *discordgo.MessageDelete) {
		cb(m)
	}

	if once {
		c.ses.AddHandlerOnce(mdcb)
	} else {
		c.ses.AddHandler(mdcb)
	}
}

// OnReady handles a Discord ``READY`` event
func (c *DiscordClient) OnReady(once bool, cb func(*discordgo.Ready)) {
	readyCb := func(_ *discordgo.Session, r *discordgo.Ready) {
		cb(r)
	}

	if once {
		c.ses.AddHandlerOnce(readyCb)
	} else {
		c.ses.AddHandler(readyCb)
	}
}

func (c *DiscordClient) OnGuildUpdate(once bool, cb func(*DiscordGuild)) {
	handleCb := func(_ *discordgo.Session, edit *discordgo.GuildUpdate) {
		res := c.Cache.UpdateGuild(edit.Guild)
		cb(res)
	}

	if once {
		c.ses.AddHandlerOnce(handleCb)
	} else {
		c.ses.AddHandler(handleCb)
	}
}

func (c *DiscordClient) handleGuildD(s *discordgo.Session, g *discordgo.GuildDelete) {
	c.Cache.DeleteGuild(g.ID)
}

func (c *DiscordClient) OnChannelCreate(once bool, cb func(*DiscordChannel)) {
	hndlerCb := func(_ *discordgo.Session, cc *discordgo.ChannelCreate) {
		ch := c.Cache.UpdateChannel(cc.Channel)
		cb(ch)
	}

	if once {
		c.ses.AddHandlerOnce(hndlerCb)
	} else {
		c.ses.AddHandler(hndlerCb)
	}
}

func (c *DiscordClient) OnChannelUpdate(once bool, cb func(*DiscordChannel)) {
	handlerCb := func(_ *discordgo.Session, cu *discordgo.ChannelUpdate) {
		ch := c.Cache.UpdateChannel(cu.Channel)
		cb(ch)
	}

	if once {
		c.ses.AddHandlerOnce(handlerCb)
	} else {
		c.ses.AddHandler(handlerCb)
	}
}

func (c *DiscordClient) OnChannelDelete(once bool, cb func(*DiscordChannel)) {
	handlerCb := func(_ *discordgo.Session, cd *discordgo.ChannelDelete) {
		ch := c.Cache.GetChannel(cd.ID)
		c.Cache.DeleteChannel(cd.ID)
		cb(ch)
	}

	if once {
		c.ses.AddHandlerOnce(handlerCb)
	} else {
		c.ses.AddHandler(handlerCb)
	}
}

func (c *DiscordClient) OnMemberAdd(once bool, cb func(*DiscordMember)) {
	handlerCb := func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		mem := c.Cache.UpdateMember(m.Member)
		cb(mem)
	}

	if once {
		c.ses.AddHandlerOnce(handlerCb)
	} else {
		c.ses.AddHandler(handlerCb)
	}
}

func (c *DiscordClient) OnMemberRemove(once bool, cb func(*DiscordMember)) {
	handlerCb := func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
		mem := c.Cache.GetMember(m.GuildID, m.User.ID)
		c.Cache.DeleteMember(m.User.ID)
		cb(mem)
	}

	if once {
		c.ses.AddHandlerOnce(handlerCb)
	} else {
		c.ses.AddHandler(handlerCb)
	}
}

func (c *DiscordClient) OnMemberUpdate(once bool, cb func(*DiscordMember)) {
	handlerCb := func(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
		mem := c.Cache.UpdateMember(m.Member)
		cb(mem)
	}

	if once {
		c.ses.AddHandlerOnce(handlerCb)
	} else {
		c.ses.AddHandler(handlerCb)
	}
}

// WithMemberChunk handles a ``GuildMembersChunk`` event.
func (c *DiscordClient) WithMemberChunk(once bool, cb func(*discordgo.GuildMembersChunk)) {
	chunkCb := func(_ *discordgo.Session, chunk *discordgo.GuildMembersChunk) {
		cb(chunk)
	}
	if once {
		c.ses.AddHandlerOnce(chunkCb)
	} else {
		c.ses.AddHandler(chunkCb)
	}
}

func (c *DiscordClient) WithGuildBanAdd(once bool, cb func(*DiscordGuildBan)) {
	guildBanAddCb := func(s *discordgo.Session, ban *discordgo.GuildBanAdd) {
		result := &DiscordGuildBan{
			User: NewDiscordUser(c, ban.User),
		}
		if cg := c.Cache.GetGuild(ban.GuildID); cg != nil {
			result.Guild = cg
		} else {
			g, _ := s.Guild(ban.GuildID)
			result.Guild = NewDiscordGuild(c, g)
		}
		cb(result)
	}

	if once {
		c.ses.AddHandlerOnce(guildBanAddCb)
	} else {
		c.ses.AddHandler(guildBanAddCb)
	}
}

func (c *DiscordClient) WithGuildBanRemove(once bool, cb func(*DiscordGuildBan)) {
	guildBanRemoveCb := func(s *discordgo.Session, ban *discordgo.GuildBanRemove) {
		result := &DiscordGuildBan{
			User: NewDiscordUser(c, ban.User),
		}
		if cg := c.Cache.GetGuild(ban.GuildID); cg != nil {
			result.Guild = cg
		} else {
			g, _ := s.Guild(ban.GuildID)
			result.Guild = NewDiscordGuild(c, g)
		}
		cb(result)
	}

	if once {
		c.ses.AddHandlerOnce(guildBanRemoveCb)
	} else {
		c.ses.AddHandler(guildBanRemoveCb)
	}
}

func (c *DiscordClient) OnMessageReactionAdd(once bool, cb func(*discordgo.MessageReactionAdd)) {
	react := func(_ *discordgo.Session, r *discordgo.MessageReactionAdd) {
		cb(r)
	}

	if once {
		c.ses.AddHandlerOnce(react)
	} else {
		c.ses.AddHandler(react)
	}
}
