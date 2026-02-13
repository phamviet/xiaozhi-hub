package ws

import (
	"github.com/lxzan/gws"
	"github.com/phamviet/xiaozhi-hub/internal/hub/ws/types"
)

func (c *Client) SendTtsStart(sampleRate int) error {
	return c.SendJSON(types.TTSMessage{
		BaseMessage: types.BaseMessage{
			Type:      types.MessageTypeTTS,
			SessionID: c.SessionID(),
		},
		State:      "start",
		SampleRate: sampleRate,
	})
}

func (c *Client) SendTtsStop() error {
	return c.SendJSON(types.TTSMessage{
		BaseMessage: types.BaseMessage{
			Type:      types.MessageTypeTTS,
			SessionID: c.SessionID(),
		},
		State: "stop",
	})
}

func (c *Client) SendTtsMessage(state string, text string) error {
	return c.SendJSON(types.TTSMessage{
		BaseMessage: types.BaseMessage{
			Type:      types.MessageTypeTTS,
			SessionID: c.SessionID(),
		},
		State: state,
		Text:  text,
	})
}

func (c *Client) SendSttMessage(text string) error {
	return c.SendJSON(types.TTSMessage{
		BaseMessage: types.BaseMessage{
			Type:      types.MessageTypeSTT,
			SessionID: c.SessionID(),
		},
		Text: text,
	})
}

func (c *Client) SendAudio(audio []byte) error {
	return c.SendMessage(gws.OpcodeBinary, audio)
}
