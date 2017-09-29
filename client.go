package dgofw

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type (
	DiscordEvent int

	DiscordCache struct {
		sync.RWMutex
		client   *DiscordClient
		users    map[string]*DiscordUser
		members  map[string]*DiscordMember
		guilds   map[string]*DiscordGuild
		channels map[string]*DiscordChannel
	}

	Handler struct {
		once    bool
		pattern string
		msgCb   func(*DiscordMessage)
		cb      interface{}
		typ     DiscordEvent
	}

	DiscordClient struct {
		sync.RWMutex
		ses          *discordgo.Session
		handlers     []*Handler
		interceptors []chan *DiscordMessage
	}

	DiscordReady struct {
		Ready *discordgo.Ready
	}
)

const (
	_ DiscordEvent = iota
	evMessage
	evGuild
	evChannel
	evMember
	evReady
)

// Cache is a global cache containing several dgofw objects
var Cache = &DiscordCache{}

// GetGuild gets a guild from the cache
func (c *DiscordCache) GetGuild(id string) *DiscordGuild {
	if g, ok := c.guilds[id]; ok {
		return g
	}
	return nil
}

// GetChannel gets a channel from the cache
func (c *DiscordCache) GetChannel(id string) *DiscordChannel {
	if c, ok := c.channels[id]; ok {
		return c
	}
	return nil
}

// GetMember gets a member from the cache
func (c *DiscordCache) GetMember(id string) *DiscordMember {
	if m, ok := c.members[id]; ok {
		return m
	}
	return nil
}

func (c *DiscordClient) initCache() {
	Cache.users = make(map[string]*DiscordUser)
	Cache.members = make(map[string]*DiscordMember)
	Cache.guilds = make(map[string]*DiscordGuild)
	Cache.channels = make(map[string]*DiscordChannel)
}

func (c *DiscordClient) initEvents() {
	c.ses.AddHandler(c.handleMessageC)
	c.ses.AddHandler(c.handleMessageE)
	c.ses.AddHandler(c.handleReady)
}

func (c *DiscordClient) intercept(timeout int, recv chan *DiscordMessage) chan struct{} {
	ret := make(chan struct{})
	go func(recv chan *DiscordMessage, closer chan struct{}) {
		ticker := time.NewTicker(time.Duration(timeout) * time.Second)
		inter := make(chan *DiscordMessage)
		c.interceptors = append(c.interceptors, inter)
		index := len(c.interceptors) - 1
		stop := false
		for !stop {
			select {
			case msg := <-inter:
				recv <- msg
			case <-ticker.C:
				stop = true
			case <-closer:
				stop = true
			default:
				stop = true
			}
		}
		ticker.Stop()
		c.Lock()
		if len(c.interceptors) > 0 {
			c.interceptors = append(c.interceptors[:index-1], c.interceptors[index:]...)
		} else {
			c.interceptors = c.interceptors[1:]
		}
		c.Unlock()
		closer <- struct{}{}
	}(recv, ret)
	return ret
}

func (c *DiscordClient) waitForMessage(timeout int, cb func(*DiscordMessage) bool) {
	reader := make(chan *DiscordMessage)
	closer := c.intercept(15, reader)
	defer close(reader)
	defer close(closer)

	for {
		select {
		case msg := <-reader:
			if cb(msg) {
				closer <- struct{}{}
				return
			}
		case <-closer:
			return
		default:
			return
		}
	}
}

// NewDiscordClient makes a new DiscordClient
//
// Automatically pre-pends 'Bot ' to the token.
func NewDiscordClient(token string) *DiscordClient {
	var err error
	result := new(DiscordClient)
	result.ses, err = discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	result.initCache()
	result.initEvents()
	return result
}

// Connect connects the client
func (c *DiscordClient) Connect() {
	c.ses.Open()
}

// Disconnect disconnects a client
func (c *DiscordClient) Disconnect() {
	c.ses.Close()
}

// SetStatus sets the ``Playing ...`` status for the bot.
func (c *DiscordClient) SetStatus(status string) {
	c.ses.UpdateStatus(0, status)
}

// TODO :: channel + guild + member event handlers
