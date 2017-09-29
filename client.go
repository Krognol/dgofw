package dgofw

import (
	"fmt"
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

	DiscordClient struct {
		sync.RWMutex
		ses          *discordgo.Session
		handlers     []*Handler
		interceptors []chan *DiscordMessage
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

func (c *DiscordCache) UpdateGuild(s *discordgo.Session, g *discordgo.Guild) {
	if gc, ok := c.guilds[g.ID]; ok {
		gc.g = g
	} else {
		c.Lock()
		c.guilds[g.ID] = NewDiscordGuild(s, g)
		c.Unlock()
	}
}

func (c *DiscordCache) DeleteGuild(id string) {
	c.Lock()
	delete(c.guilds, id)
	c.Unlock()
}

// GetChannel gets a channel from the cache
func (c *DiscordCache) GetChannel(id string) *DiscordChannel {
	if c, ok := c.channels[id]; ok {
		return c
	}
	return nil
}

func (c *DiscordCache) UpdateChannel(s *discordgo.Session, ch *discordgo.Channel) {
	if cc, ok := c.channels[ch.ID]; ok {
		cc.c = ch
	} else {
		c.Lock()
		c.channels[ch.ID] = NewDiscordChannel(s, ch)
		c.Unlock()
	}
}

func (c *DiscordCache) DeleteChannel(id string) {
	c.Lock()
	delete(c.channels, id)
	c.Unlock()
}

// GetMember gets a member from the cache
func (c *DiscordCache) GetMember(id string) *DiscordMember {
	if m, ok := c.members[id]; ok {
		return m
	}
	return nil
}

func (c *DiscordCache) UpdateMember(s *discordgo.Session, m *discordgo.Member) {
	if cm, ok := c.members[m.User.ID]; ok {
		cm.m = m
	} else {
		c.Lock()
		c.members[m.User.ID] = NewDiscordMember(s, m)
		c.Unlock()
	}
}

func (c *DiscordCache) DeleteMember(id string) {
	c.Lock()
	delete(c.members, id)
	c.Unlock()
}

func (c *DiscordClient) initCache() {
	Cache.client = c
	Cache.users = make(map[string]*DiscordUser)
	Cache.members = make(map[string]*DiscordMember)
	Cache.guilds = make(map[string]*DiscordGuild)
	Cache.channels = make(map[string]*DiscordChannel)
}

func (c *DiscordClient) initEvents() {
	// Message Event Handlers
	c.ses.AddHandler(c.handleMessageC)
	c.ses.AddHandler(c.handleMessageE)
	//c.ses.AddHandler(c.handleMessageD)

	// Guild Event Handlers

	// We ignore GUILD_CREATE
	c.ses.AddHandler(c.handleGuildE)
	c.ses.AddHandler(c.handleGuildD)
	c.ses.AddHandler(c.handleReady)

	c.ses.AddHandler(c.handleMemberA)
	c.ses.AddHandler(c.handleMemberE)
	c.ses.AddHandler(c.handleMemberD)
}

func (c *DiscordClient) intercept(timeout int, closer chan struct{}, onLimit func()) (reader chan *DiscordMessage) {
	reader = make(chan *DiscordMessage)
	c.interceptors = append(c.interceptors, reader)

	go func() {
		ticker := time.NewTicker(time.Duration(timeout) * time.Second)
		index := len(c.interceptors) - 1
		stop := false

		for !stop {
			select {
			case <-ticker.C:
				stop = true
				if onLimit != nil {
					onLimit()
				}
			case <-closer:
				stop = true
				close(closer)
			}
		}

		ticker.Stop()
		c.Lock()
		if index > 0 {
			c.interceptors = append(c.interceptors[:index], c.interceptors[index+1:]...)
		} else {
			c.interceptors = c.interceptors[1:]
		}
		c.Unlock()
	}()
	return
}

func (c *DiscordClient) waitForMessage(timeout int, cb func(*DiscordMessage) bool, onLimit func()) {
	closer := make(chan struct{})
	reader := c.intercept(15, closer, onLimit)

	for msg := range reader {
		if cb(msg) {
			closer <- struct{}{}
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
	result.interceptors = make([]chan *DiscordMessage, 0)
	result.initCache()
	result.initEvents()
	return result
}

// Connect connects the client
func (c *DiscordClient) Connect() {
	err := c.ses.Open()
	if err != nil {
		fmt.Println(err)
	}
}

// Disconnect disconnects a client
func (c *DiscordClient) Disconnect() {
	c.ses.Close()
}

// SetStatus sets the ``Playing ...`` status for the bot.
func (c *DiscordClient) SetStatus(status string) {
	c.ses.UpdateStatus(0, status)
}
