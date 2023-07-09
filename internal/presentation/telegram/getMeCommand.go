package telegram

import (
	"fmt"
	"strings"

	"github.com/anonyindian/gotgproto/ext"
)

func (r *Presentation) getMeCommandHandler(ctx *ext.Context, update *ext.Update) error {
	var textBuilder strings.Builder
	const requestedUserTmp = "Requested user: \n" +
		"id: <code>%d</code>\n" +
		"username: @%s\n\n"
	const currentChatTmp = "Current chat: \n" +
		"id: <code>%d</code>"
	user := update.EffectiveUser()
	textBuilder.WriteString(fmt.Sprintf(requestedUserTmp, user.ID, user.Username))

	chat := update.EffectiveChat()
	textBuilder.WriteString(fmt.Sprintf(currentChatTmp, chat.GetID()))

	_, err := ctx.Reply(update, textBuilder.String(), nil)
	return err
}
