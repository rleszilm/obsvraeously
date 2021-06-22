package avrae

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/rleszilm/genms/service"
)

type Service struct {
	service.Dependencies

	auth  string
	rolls chan *Roll
	dg    *discordgo.Session
}

func NewService(auth string) *Service {
	return &Service{
		auth: auth,
	}
}

func (s *Service) Initialize(ctx context.Context) error {
	s.rolls = make(chan *Roll, 32)

	dg, err := discordgo.New("Bot " + s.auth)
	if err != nil {
		return err
	}
	s.dg = dg
	s.dg.AddHandler(NewCreateHandler(s.rolls))
	s.dg.AddHandler(NewUpdateHandler(s.rolls))

	s.dg.Identify.Intents = discordgo.IntentsGuildMessages

	if err = dg.Open(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	if err := s.dg.Close(); err != nil {
		return err
	}
	close(s.rolls)

	return nil
}

func (s *Service) NameOf() string {
	return "discord"
}

func (s *Service) String() string {
	return s.NameOf()
}

func (s *Service) Rolls() <-chan *Roll {
	return s.rolls
}
