package dgofw

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type (
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
		Cache            *DiscordCache
		ses              *discordgo.Session
		handlers         []*MsgHandler
		interceptors     []chan *DiscordMessage
		VoiceConnections []*DiscordVoiceConnection
	}
)

// GetGuild gets a guild from the cache
func (c *DiscordCache) GetGuild(id string) *DiscordGuild {
	if g, ok := c.guilds[id]; ok {
		return g
	}

	if g, err := c.client.ses.State.Guild(id); err == nil {
		return NewDiscordGuild(c.client, g)
	}

	g, err := c.client.ses.Guild(id)
	if err != nil {
		return nil
	}

	return NewDiscordGuild(c.client, g)
}

func (c *DiscordCache) UpdateGuild(g *discordgo.Guild) *DiscordGuild {
	var result *DiscordGuild
	if gc, ok := c.guilds[g.ID]; ok {
		gc.g = g
		result = gc
	} else {
		result = NewDiscordGuild(c.client, g)
		c.Lock()
		c.guilds[g.ID] = result
		c.Unlock()
	}
	return result
}

func (c *DiscordCache) DeleteGuild(id string) {
	c.Lock()
	if _, ok := c.guilds[id]; ok {
		delete(c.guilds, id)
	}
	c.Unlock()
}

// GetChannel gets a channel from the cache
func (c *DiscordCache) GetChannel(id string) *DiscordChannel {
	if ch, ok := c.channels[id]; ok {
		return ch
	}

	if ch, err := c.client.ses.State.Channel(id); err == nil {
		return NewDiscordChannel(c.client, ch)
	}

	ch, err := c.client.ses.Channel(id)
	if err != nil {
		return nil
	}

	return NewDiscordChannel(c.client, ch)
}

func (c *DiscordCache) UpdateChannel(ch *discordgo.Channel) *DiscordChannel {
	var result *DiscordChannel
	if cc, ok := c.channels[ch.ID]; ok {
		cc.c = ch
		result = cc
	} else {
		result = NewDiscordChannel(c.client, ch)
		c.Lock()
		c.channels[ch.ID] = result
		c.Unlock()
	}
	return result
}

func (c *DiscordCache) DeleteChannel(id string) {
	c.Lock()
	if _, ok := c.channels[id]; ok {
		delete(c.channels, id)
	}
	c.Unlock()
}

// GetMember gets a member from the cache
func (c *DiscordCache) GetMember(guild, id string) *DiscordMember {
	if m, ok := c.members[id]; ok {
		return m
	}

	if m, err := c.client.ses.State.Member(guild, id); err == nil {
		return NewDiscordMember(c.client, m)
	}

	m, err := c.client.ses.GuildMember(guild, id)
	if err != nil {
		return nil
	}

	return NewDiscordMember(c.client, m)
}

func (c *DiscordCache) UpdateMember(m *discordgo.Member) *DiscordMember {
	var result *DiscordMember
	if cm, ok := c.members[m.User.ID]; ok {
		cm.m = m
		result = cm
	} else {
		result = NewDiscordMember(c.client, m)
		c.Lock()
		c.members[m.User.ID] = result
		c.Unlock()
	}
	return result
}

func (c *DiscordCache) DeleteMember(id string) {
	c.Lock()
	if _, ok := c.members[id]; ok {
		delete(c.members, id)
	}
	c.Unlock()
}

func (c *DiscordClient) initCache() {
	c.Cache = &DiscordCache{
		client:   c,
		users:    make(map[string]*DiscordUser),
		members:  make(map[string]*DiscordMember),
		guilds:   make(map[string]*DiscordGuild),
		channels: make(map[string]*DiscordChannel),
	}
}

func (c *DiscordClient) initEvents() {
	// Message Event Handlers
	c.ses.AddHandler(c.handleMessageC)
	c.ses.AddHandler(c.handleMessageE)

	// Guild Event Handlers

	// We ignore GUILD_CREATE
	c.ses.AddHandler(c.handleGuildD)
}

func (c *DiscordClient) intercept(timeout int, closer chan struct{}, onLimit func()) (reader chan *DiscordMessage) {
	reader = make(chan *DiscordMessage)

	c.Lock()
	c.interceptors = append(c.interceptors, reader)
	c.Unlock()

	go func(c *DiscordClient, timeout int, closer chan struct{}, onLimit func()) {
		index := len(c.interceptors) - 1
		ticker := time.NewTicker(time.Second * time.Duration(timeout))

		select {
		case <-ticker.C:
			if onLimit != nil {
				onLimit()
			}
		case <-closer:
			break
		}

		ticker.Stop()

		c.Lock()
		if len(c.interceptors) > 0 {
			copy(c.interceptors[index:], c.interceptors[index+1:])
			c.interceptors[len(c.interceptors)-1] = nil
			c.interceptors = c.interceptors[:len(c.interceptors)-1]
		}
		c.Unlock()

		close(closer)
	}(c, timeout, closer, onLimit)
	return
}

func (c *DiscordClient) waitForMessage(timeout int, cb func(*DiscordMessage) bool, onLimit func()) {
	closer := make(chan struct{})
	reader := c.intercept(timeout, closer, onLimit)
	defer close(reader)

	for msg := range reader {
		if cb(msg) {
			closer <- struct{}{}
			return
		}
	}
}

func (c *DiscordClient) interceptForever(closer chan bool) (reader chan *DiscordMessage) {
	reader = make(chan *DiscordMessage)

	c.RLock()
	c.interceptors = append(c.interceptors, reader)
	c.RUnlock()

	go func() {
		index := len(c.interceptors) - 1
		<-closer

		c.Lock()
		if index > 0 {
			c.interceptors = append(c.interceptors[:index], c.interceptors[index+1:]...)
		}
		c.Unlock()
		close(closer)
	}()
	return
}

func (c *DiscordClient) waitForever(cb func(*DiscordMessage)) chan bool {
	closer := make(chan bool)
	reader := c.interceptForever(closer)
	go func(cb func(*DiscordMessage), closer chan bool) {
		defer close(closer)
		defer close(reader)
		for {
			select {
			case msg := <-reader:
				cb(msg)
			case t := <-closer:
				if t {
					return
				}
			}
		}
	}(cb, closer)
	return closer
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

func (c *DiscordClient) Channel(id string) *DiscordChannel {
	if ch := c.Cache.GetChannel(id); ch != nil {
		return ch
	}
	return nil
}

func (c *DiscordClient) Send(channel, msg string) *DiscordMessage {
	m, err := c.ses.ChannelMessageSend(channel, msg)
	if err != nil {
		return nil
	}

	return NewDiscordMessage(c, m)
}

func (c *DiscordClient) SendEmbed(channel string, embed *discordgo.MessageEmbed) (err error) {
	_, err = c.ses.ChannelMessageSendEmbed(channel, embed)
	return
}
