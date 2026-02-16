package ws

import (
	"log"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func (c *Client) initInternalTools() {
	var tools []ai.ToolRef
	exitTool := genkit.DefineTool(c.g, "exit_intent", "Use this when user want to stop the conversation",
		func(ctx *ai.ToolContext, input any) (any, error) {
			log.Println("exit_intent called")
			c.exitIntentCalled = true
			return nil, nil
		},
	)

	tools = append(tools, exitTool)
	c.tools = tools
}
