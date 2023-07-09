package telegram

import (
	"errors"
	"fmt"

	"github.com/anonyindian/gotgproto/dispatcher"
	"github.com/anonyindian/gotgproto/ext"
	"github.com/rs/zerolog/log"

	"github.com/anonyindian/gotgproto"
	"github.com/anonyindian/gotgproto/dispatcher/handlers"
	"github.com/anonyindian/gotgproto/sessionMaker"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/teadove/goteleout/internal/service/client"
	"github.com/teadove/goteleout/internal/service/storage"
	"github.com/teadove/goteleout/internal/utils"
)

var BadUpdate = errors.New("bad update")

type Presentation struct {
	telegramClient  *telegram.Client
	telegramSender  *message.Sender
	telegramApi     *tg.Client
	telegramManager *peers.Manager
	protoClient     *gotgproto.Client

	storage       storage.Interface
	clientService *client.Service

	logErrorToSelf bool
}

func MustNewTelegramPresentation(
	clientService *client.Service,
	telegramAppID int,
	telegramAppHash string,
	telegramPhoneNumber string,
	telegramSessionStorageFullPath string,
	storage storage.Interface,
	logErrorToSelf bool,
) Presentation {
	protoClient, err := gotgproto.NewClient(telegramAppID, telegramAppHash, gotgproto.ClientType{
		Phone: telegramPhoneNumber,
	}, &gotgproto.ClientOpts{
		DisableCopyright: true,
		Session: sessionMaker.NewSession(
			telegramSessionStorageFullPath,
			sessionMaker.Session,
		),
	})
	utils.Check(err)

	api := protoClient.API()

	presentation := Presentation{
		clientService:   clientService,
		storage:         storage,
		protoClient:     protoClient,
		telegramApi:     api,
		telegramSender:  message.NewSender(api),
		telegramManager: peers.Options{}.Build(api),
		logErrorToSelf:  logErrorToSelf,
	}

	protoClient.Dispatcher.AddHandler(handlers.NewCommand("echo", presentation.echoCommandHandler))
	protoClient.Dispatcher.AddHandler(handlers.NewCommand("help", presentation.helpCommandHandler))
	protoClient.Dispatcher.AddHandler(
		handlers.NewCommand("get_me", presentation.getMeCommandHandler),
	)
	protoClient.Dispatcher.AddHandler(handlers.NewCommand("ping", presentation.pingCommandHandler))
	protoClient.Dispatcher.AddHandler(
		handlers.NewCommand("spam_reaction", presentation.spamReactionCommandHandler),
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.spamReactionMessageHandler,
			Filters:       nil,
			UpdateFilters: nil,
			Outgoing:      true,
		},
	)
	dp, ok := protoClient.Dispatcher.(*dispatcher.NativeDispatcher)
	if !ok {
		utils.Check(errors.New("can only work with NativeDispatcher"))
	}
	dp.Error = presentation.errorHandler

	return presentation
}

func (r *Presentation) errorHandler(
	ctx *ext.Context,
	update *ext.Update,
	errorString string,
) error {
	log.Error().
		Stack().
		Err(errors.New(errorString)).
		Str("status", "error.while.processing.update").
		Interface("update", update).
		Send()
	if r.logErrorToSelf {
		_, err := ctx.SendMessage(ctx.Self.ID, &tg.MessagesSendMessageRequest{
			Silent:  true,
			Message: fmt.Sprintf("Error occured while processing update:\n\n%s", errorString),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Presentation) Run() error {
	ctx := r.protoClient.CreateContext()
	_, err := ctx.SendMessage(
		ctx.Self.ID,
		&tg.MessagesSendMessageRequest{Message: "Fun telegram initialized!"},
	)
	if err != nil {
		return err
	}

	err = r.protoClient.Idle()
	return err
}
