package macaroons

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

type MacLogger struct {
	zerolog.Logger
}

var _ bakery.Logger = (*MacLogger)(nil)

// Infof is part of the macaroon-bakery Logger interface
func (m *MacLogger) Infof(ctx context.Context, f string, args ...interface{}) {
	m.Info().Msg(fmt.Sprintf(f, args))
}

// Debugf is part of the macaroon-bakery Logger interface
func (m *MacLogger) Debugf(ctx context.Context, f string, args ...interface{}) {
	m.Debug().Msg(fmt.Sprintf(f, args))
}
