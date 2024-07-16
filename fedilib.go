package fedilib

import (
	"context"
	"errors"
	"log"

	"github.com/mattn/go-mastodon"
)

const (
	DefaultUserAgent = "fedilib/0.01 (2024-03-23)"
)

type Application interface {
	HandleNotification(n *mastodon.Notification)
}

type Config struct {
	// https://botsin.space
	Server string `yaml:"server,omitempty" json:"server,omitempty"`
	// Client key: kZoi323…
	ClientID string `yaml:"client_id,omitempty" json:"client_id,omitempty"`
	// Client secret: ose…
	ClientSecret string `yaml:"client_secret,omitempty" json:"client_secret,omitempty"`
	// Application name: fedilpd
	ClientName string `yaml:"client_name,omitempty" json:"client_name,omitempty"`
	// Scopes: read write follow
	Scopes string `yaml:"scopes,omitempty" json:"scopes,omitempty"`
	// Application website: https://berlin.de/presse
	Website string `yaml:"website,omitempty" json:"website,omitempty"`
	// Redirect URI: urn:ietf:wg:oauth:2.0:oob
	RedirectURI string `yaml:"redirect_uri,omitempty" json:"redirect_uri,omitempty"`
	// Your access token: Rdn…
	Token     string `yaml:"token,omitempty" json:"token,omitempty"`
	UserAgent string `yaml:"user_agent,omitempty" json:"user_agent,omitempty"`
}

type Fedi struct {
	client *mastodon.Client
	ctx    context.Context
	log    *log.Logger
	self   *mastodon.Account
	app    Application
}

func (s *Fedi) Init(cfg *Config, app Application, log *log.Logger) {

	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}

	c := &mastodon.Client{
		Config: &mastodon.Config{
			Server:       cfg.Server,
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			AccessToken:  cfg.Token,
		},
		UserAgent: cfg.UserAgent,
	}

	s.client = c
	s.ctx = context.Background()
	s.log = log
	s.app = app
}

func (s *Fedi) Client() *mastodon.Client {
	return s.client
}

func (s *Fedi) Ctx() context.Context {
	return s.ctx
}

func (s *Fedi) Log() *log.Logger {
	return s.log
}

func (s *Fedi) WSClient() *mastodon.WSClient {
	return s.Client().NewWSClient()
}

func (s *Fedi) ProcessNotifications() {
	pg := mastodon.Pagination{
		Limit: 40,
	}
	if nl, err := s.Client().GetNotifications(s.Ctx(), &pg); err != nil {
		s.Log().Fatal(err)
	} else {
		for _, n := range nl {
			s.app.HandleNotification(n)
		}
	}

	s.Client().ClearNotifications(s.Ctx())
}

func (s *Fedi) WatchNotifications() {
	if evc, err := s.WSClient().StreamingWSUser(s.Ctx()); err != nil {
		s.Log().Fatal(err)
	} else {
		for e := range evc {
			switch me := e.(type) {
			// case *mastodon.UpdateEvent:
			// 	s.Log().Print(me.Status)
			case *mastodon.NotificationEvent:
				s.app.HandleNotification(me.Notification)
			default:
				s.Log().Printf("Unhandled event %#v", me)
			}
		}
	}
}

func (s *Fedi) IsFollower(account *mastodon.Account) error {
	pg := &mastodon.Pagination{
		Limit: 40,
	}
	if la, err := s.Client().GetAccountFollowers(s.Ctx(), s.CurrentAccount().ID, pg); err != nil {
		return err
	} else {
		for _, l := range la {
			if l.ID == account.ID {
				return nil
			}
		}
	}
	return errors.New("ignore from non follower " + account.Acct)
}

func (s *Fedi) CurrentAccount() *mastodon.Account {
	if s.self == nil {
		if self, err := s.Client().GetAccountCurrentUser(s.Ctx()); err != nil {
			s.Log().Fatal(err)
		} else {
			s.self = self
		}
	}
	return s.self
}

func (s *Fedi) GetList(title string) *mastodon.List {
	if ll, err := s.Client().GetLists(s.Ctx()); err != nil {
		s.Log().Fatal(err)
	} else {
		for _, l := range ll {
			if l.Title == title {
				return l
			}
		}
	}

	l, err := s.Client().CreateList(s.Ctx(), title)
	if err != nil {
		s.Log().Fatal(err)
	}
	return l
}

func (s *Fedi) MarkAccount(account *mastodon.Account, mark string) error {
	// remove account from all other lists
	if ll, err := s.Client().GetAccountLists(s.Ctx(), s.CurrentAccount().ID); err == nil {
		for _, l := range ll {
			if l.Title == mark {
				continue
			}
			if err = s.Client().RemoveFromList(s.Ctx(), l.ID, account.ID); err != nil {
				s.Log().Print(err)
			}
		}
	} else {
		s.Log().Print(err)
	}

	list := s.GetList(mark)
	err := s.Client().AddToList(s.Ctx(), list.ID, account.ID)
	return err
}
