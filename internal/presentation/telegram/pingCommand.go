package telegram

import (
	"fmt"
	"strings"

	"github.com/anonyindian/gotgproto/ext"
	"github.com/anonyindian/gotgproto/types"
	"github.com/rs/zerolog/log"

	"github.com/gotd/td/telegram/peers/members"
)

func (r *Presentation) pingCommandHandler(ctx *ext.Context, update *ext.Update) error {
	const maxCount = 15
	count := 0
	requestedUser := update.EffectiveUser()

	var textBuilder strings.Builder
	textBuilder.Grow(100)
	textBuilder.WriteString(fmt.Sprintf("Ping requested by %s\n\n", requestedUser.Username))
	var mentionBuilder strings.Builder
	compileMention := func(p members.Member) error {
		user := p.User()
		_, isBot := user.ToBot()
		if isBot {
			return nil
		}
		username, ok := user.Username()
		if !ok {
			return nil
		}
		count += 1
		mentionBuilder.WriteString(fmt.Sprintf("@%s\n", username))
		return nil
	}

	switch t := update.EffectiveChat().(type) {
	case *types.Chat:
		chat := r.telegramManager.Chat(t.Raw())
		chatMembers := members.Chat(chat)
		err := chatMembers.ForEach(ctx, compileMention)
		if err != nil {
			return err
		}
	case *types.Channel:
		chat := r.telegramManager.Channel(t.Raw())
		chatMembers := members.Channel(chat)
		err := chatMembers.ForEach(ctx, compileMention)
		if err != nil {
			return err
		}
	default:
		_, err := ctx.Reply(update, "Err: this command work only in chats", nil)
		return err
	}

	if mentionBuilder.String() == "" {
		log.Warn().Str("status", "no.users.were.mentioned").Send()
		return nil
	}
	if count > maxCount {
		log.Warn().Str("status", "max.count.exceeded").Send()
		_, err := ctx.Reply(
			update,
			fmt.Sprintf("Max user count exceeded, count: %d, maxCount: %d", count, maxCount),
			nil,
		)
		return err
	}
	textBuilder.WriteString(mentionBuilder.String())
	_, err := ctx.Reply(update, textBuilder.String(), nil)
	return err
}
