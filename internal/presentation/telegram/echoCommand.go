package telegram

import "github.com/anonyindian/gotgproto/ext"

func (r *Presentation) echoCommandHandler(ctx *ext.Context, update *ext.Update) error {
	_, err := ctx.Reply(update, update.EffectiveMessage.Message.Message, nil)
	return err
}
