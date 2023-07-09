package utils

import (
	"github.com/anonyindian/gotgproto/types"
	"github.com/gotd/td/tg"
)

func GetChatFromEffectiveChat(effectiveChat types.EffectiveChat) (int64, tg.InputPeerClass) {
	switch t := effectiveChat.(type) {
	case *types.Chat, *types.User, *types.Channel:
		return t.GetID(), t.GetInputPeer()
	default:
		return 0, &tg.InputPeerEmpty{}
	}
}
