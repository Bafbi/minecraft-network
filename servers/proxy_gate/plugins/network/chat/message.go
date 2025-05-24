package chat

import (
	// Assuming util.Join is available at this path
	. "github.com/bafbi/minecraft-network/servers/proxy_gate/util"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
)

// formatChatMessage formats a chat message with username and message content.
func formatChatMessage(username, msg string) *c.Text {
	name := &c.Text{
		Content: username,
		S:       c.Style{Color: color.Gold},
	}
	separator := &c.Text{
		Content: ": ",
		S:       c.Style{Color: color.White},
	}
	message := &c.Text{
		Content: msg,
		S:       c.Style{Color: color.White, Italic: c.True},
	}
	return Join(name, separator, message)
}
