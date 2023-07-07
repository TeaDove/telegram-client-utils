package telegram

import (
	"context"
	"errors"
	"github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
	"strings"
	"time"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/service/client"
	"golang.org/x/time/rate"
)

type Presentation struct {
	telegramClient     *telegram.Client
	telegramDispatcher *tg.UpdateDispatcher
	telegramSender     *message.Sender
	telegramApi        *tg.Client
	telegramManager    *peers.Manager
	commandHandler     map[string]commandProcessor
	waiter             *floodwait.Waiter

	storage       storage.Interface
	clientService *client.Service
}

type commandProcessor func(ctx context.Context, entities *tg.Entities, update message.AnswerableMessageUpdate, m *tg.Message) error

func MustNewTelegramPresentation(
	clientService *client.Service,
	telegramAppID int,
	telegramAppHash string,
	telegramSessionStorageFullPath string,
	storage storage.Interface,
) Presentation {
	// https://core.telegram.org/api/obtaining_api_id

	sessionStorage := telegram.FileSessionStorage{Path: telegramSessionStorageFullPath}
	updater := tg.NewUpdateDispatcher()

	waiter := floodwait.NewWaiter().
		WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
			log.Warn().Str("status", "flood.waiting").Dur("wait", wait.Duration).Send()
		})

	telegramClient := telegram.NewClient(
		telegramAppID,
		telegramAppHash,
		telegram.Options{
			SessionStorage: &sessionStorage,
			UpdateHandler:  updater,
			Middlewares: []telegram.Middleware{
				ratelimit.New(rate.Every(time.Millisecond*100), 5),
				waiter,
			},
		},
	)
	api := telegramClient.API()
	sender := message.NewSender(api)

	presentation := Presentation{
		telegramClient:     telegramClient,
		clientService:      clientService,
		telegramDispatcher: &updater,
		telegramApi:        api,
		telegramSender:     sender,
		waiter:             waiter,
		storage:            storage,
	}
	presentation.commandHandler = map[string]commandProcessor{
		"ping":         presentation.pingCommandHandler,
		"help":         presentation.helpCommandHandler,
		"getMe":        presentation.getMeCommandHandler,
		"spamReaction": presentation.spamReactionCommandHandler,
	}
	presentation.telegramManager = peers.Options{}.Build(api)

	return presentation
}

var BadUpdate = errors.New("bad update")

func (r *Presentation) login(ctx context.Context) error {
	flow := auth.NewFlow(utils.TerminalAuth{}, auth.SendCodeOptions{})
	status, err := r.telegramClient.Auth().Status(ctx)
	if !status.Authorized {
		log.Info().Str("status", "authorizing").Send()
		err = r.telegramClient.Auth().IfNecessary(ctx, flow)
	}
	if err != nil {
		return errors.Join(err, errors.New("error while authenticating"))
	}

	_, err = r.telegramSender.Self().Text(ctx, "Telegram client initialized")
	if err != nil {
		return err
	}

	return nil
}

func (r *Presentation) routeMessage(
	ctx context.Context,
	entities *tg.Entities,
	update message.AnswerableMessageUpdate,
) error {
	m, ok := update.GetMessage().(*tg.Message)
	if !ok {
		return BadUpdate
	}

	if m.Post {
		return nil
	}
	log.Debug().Str("status", "message.got").Str("text", m.Message).Interface("message", m).Send()

	fields := strings.Fields(m.Message)
	if len(fields) == 0 {
		return nil
	}

	err := r.spamReactionMessageHandler(ctx, entities, update, m)
	if err != nil {
		return err
	}

	firstMessage := fields[0]
	const commandPrefix = '!'
	if len(firstMessage) < 1 {
		return nil
	}
	if firstMessage[0] != commandPrefix {
		return nil
	}
	command := firstMessage[1:]
	handler, ok := r.commandHandler[command]
	if !ok {
		log.Warn().
			Str("status", "unknown.command").
			Str("command", command).
			Str("text", m.Message).
			Send()
		return nil
	}
	log.Info().Str("status", "command.got").Str("command", command).Send()
	t0 := time.Now()
	err = handler(ctx, entities, update, m)
	if err != nil {
		return err
	}
	log.Info().Str("status", "ok").Dur("dur", time.Now().Sub(t0)).Send()
	return nil
}

func (r *Presentation) Run() error {
	return r.waiter.Run(context.Background(), func(ctx context.Context) error {
		return r.telegramClient.Run(ctx, func(ctx context.Context) error {
			err := r.login(ctx)
			if err != nil {
				return err
			}

			r.telegramDispatcher.OnNewChannelMessage(
				func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
					err = r.routeMessage(ctx, &entities, update)
					if err != nil {
						log.Error().Stack().Err(err).Str("status", "error.while.processing.request").Send()
					}
					return err
				},
			)

			r.telegramDispatcher.OnNewMessage(
				func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
					err = r.routeMessage(ctx, &entities, update)
					if err != nil {
						log.Error().Stack().Err(err).Str("status", "error.while.processing.request").Send()
					}
					return err
				},
			)
			err = telegram.RunUntilCanceled(context.Background(), r.telegramClient)
			if err != nil {
				return err
			}

			return nil
		})
	})
}
