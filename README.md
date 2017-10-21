# dgofw

Some weird [discordgo](https://github.com/bwmarrin/discordgo) framework.

# Example

```go
package main

import (
    "fmt"
    "github.com/Krognol/dgofw"
)

func main() {
    client := dgofw.NewDiscordClient("Some token")

    // Register message handlers
    client.OnMessage("some pattern", false, func(m *dgofw.DiscordMessage){
        m.Reply("some reply")
    })

    // Register message handler once
    client.OnMessage("some other pattern", true, func(m *dgofw.DiscordMessage){
        m.Reply("some other reply")
    })

    // Message handler with args
    client.OnMessage("prefix {arg1} {arg2}", false, func(m *dgofw.DiscordMessage){
        fmt.Printf("Message arg1:%s, arg2:%s", m.Arg("arg1"), m.Arg("arg2"))
    })

    // Message handler waiting for a reply
    client.OnMessage("some pattern", false, func(m *dgofw.DiscordMessage) bool {
        m2 := m.Reply("Option 1 or 2?")

        // Wait for 15 seconds for a reply from the user
        // If there's no reply, delete the 2 messages
        m.WaitForMessage(15, func(itr *dgofw.DiscordMessage){
            if itr.Author.ID() == m.Author.ID() && itr.ChannelID() == m.ChannelID() {
                if m.Content() == "1" {
                    // do something
                } else {
                    // do something else
                }
                
                // stop waiting for reply
                return true
            }

            // continue waiting for reply
            return false
        }, /*nil or a func()*/func(){
            m.Delete()
            m2.Delete()
        })
    })

    closer := make(chan os.Signal, 1)
    signal.Notify(closer, os.Interrupt, os.Kill)

    <-closer
    client.Disconnect()
    os.Exit(0)
}
```

Most other events are also available.