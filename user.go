package dgofw

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordUser struct {
	s *discordgo.Session
	u *discordgo.User
}

func NewDiscordUser(s *discordgo.Session, u *discordgo.User) *DiscordUser {
	return &DiscordUser{
		s: s,
		u: u,
	}
}

func (u *DiscordUser) Avatar() string {
	return u.u.AvatarURL("256")
}

func (u *DiscordUser) ID() string {
	return u.u.ID
}

func (u *DiscordUser) Mention() string {
	return u.u.Mention()
}

func (u *DiscordUser) Username() string {
	return u.u.Username
}

func (u *DiscordUser) Discriminator() string {
	return u.u.Discriminator
}

func (u *DiscordUser) Timestamp() string {
	id, _ := strconv.ParseInt(u.ID(), 10, 64)
	ms := (id >> 22) + 1420070400000
	t := time.Unix(ms/1000, 0)
	return t.Format("2006-01-02 15:04:05")
}

func (u *DiscordUser) Bot() bool {
	return u.u.Bot
}

func (u *DiscordUser) Verified() bool {
	return u.u.Verified
}

func (u *DiscordUser) AsMember(guild string) *DiscordMember {
	if mem := Cache.GetMember(u.ID()); mem != nil {
		return mem
	}

	m, err := u.s.GuildMember(guild, u.ID())
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return NewDiscordMember(u.s, m)
}
