package dgofw

import (
	"fmt"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

type DiscordVoiceConnection struct {
	dv   *discordgo.VoiceConnection
	Send chan []byte
}

func (c *DiscordClient) NewVoiceConnection(guild, channel string) *DiscordVoiceConnection {
	dv, err := c.ses.ChannelVoiceJoin(guild, channel, false, false)
	if err != nil {
		return nil
	}
	result := &DiscordVoiceConnection{
		dv:   dv,
		Send: dv.OpusSend,
	}
	c.VoiceConnections = append(c.VoiceConnections, result)
	return result
}

func (vc *DiscordVoiceConnection) Speaking(isSpeaking bool) {
	err := vc.dv.Speaking(isSpeaking)
	if err != nil {
		fmt.Println(err)
	}
}

func (vc *DiscordVoiceConnection) ChangeChannel(channel string) {
	vc.dv.ChangeChannel(channel, false, false)
}

func (vc *DiscordVoiceConnection) Play(filename string, stop <-chan bool) {
	dgvoice.PlayAudioFile(vc.dv, filename, stop)
}

func (vc *DiscordVoiceConnection) Leave() {
	vc.dv.Disconnect()
}
