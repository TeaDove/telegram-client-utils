package telegram

import (
	"github.com/anonyindian/gotgproto/ext"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (r *Presentation) helpCommandHandler(ctx *ext.Context, update *ext.Update) error {
	const helpMessage = "Available commands:\n\n" +
		"<code>!help</code> - get this message\n" +
		"<code>!ping</code> - ping all users\n" +
		"<code>!get_me</code> - get id, username of requested user and group\n" +
		"<code>!spam_reaction [stop]</code> - if replied to message with reaction, will spam this reaction to replied user. " +
		"Send with stop, to stop it"

	_, err := ctx.Reply(
		update,
		[]styling.StyledTextOption{html.String(func(_ int64) (tg.InputUserClass, error) {
			return ctx.Self.AsInput(), nil
		}, helpMessage)},
		nil,
	)
	return err
}
