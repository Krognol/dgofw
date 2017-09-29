package dgofw

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (c *DiscordClient) handleMessageC(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}
	vals := strings.Fields(m.Content)

	for i, handler := range c.handlers {
		if handler.typ == evMessage {
			if len(vals) > 0 || len(m.Content) >= len(handler.pattern) {
				if m.Content == handler.pattern || strings.HasPrefix(handler.pattern, m.Content[0:len(vals[0])]) {
					msg := NewDiscordMessage(s, m.Message)
					keys := strings.Fields(handler.pattern)

					if len(keys) > 0 && len(vals) > 0 {
						msg.Pairs(keys, vals)
					}

					if len(c.interceptors) > 0 {
						go func() {
							for _, inter := range c.interceptors {
								inter <- msg
							}
						}()
					}

					go handler.msgCb(msg)
					if handler.once {
						c.RLock()
						if i > 0 {
							c.handlers = append(c.handlers[:i-1], c.handlers[i:]...)
						} else {
							c.handlers = c.handlers[1:]
						}
						c.RUnlock()
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
		msgCb:   cb,
		once:    once,
		pattern: pattern,
		typ:     evMessage,
	})
}

// OnReady handles a Discord ``READY`` event
func (c *DiscordClient) OnReady(once bool, cb func(*DiscordReady)) {
	c.handlers = append(c.handlers, &Handler{
		cb:   cb,
		once: once,
		typ:  evReady,
	})
}

func (c *DiscordClient) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	for i, h := range c.handlers {
		if h.typ == evReady {
			ready := &DiscordReady{Ready: r}
			go func(cb interface{}, r *DiscordReady) {
				cb.(func(*DiscordReady))(r)
			}(h.cb, ready)
			if h.once {
				c.RLock()
				if i > 0 {
					c.handlers = append(c.handlers[:i-1], c.handlers[i:]...)
				} else {
					c.handlers = c.handlers[1:]
				}
				c.RUnlock()
			}
		}
	}
}
