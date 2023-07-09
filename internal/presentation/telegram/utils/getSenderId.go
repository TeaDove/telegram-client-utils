package utils

import (
	"errors"

	"github.com/anonyindian/gotgproto/types"
	"github.com/gotd/td/tg"
)

func GetSenderId(m *types.Message) (int64, error) {
	peer, ok := m.GetFromID()
	if !ok {
		peer = m.PeerID
	}

	switch t := peer.(type) {
	case *tg.PeerUser:
		return t.UserID, nil
	default:
		return 0, errors.New("invalid peer")
	}
}
