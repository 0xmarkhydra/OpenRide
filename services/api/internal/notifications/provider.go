package notifications

import (
	"context"
	"errors"
)

var ErrProviderDisabled = errors.New("push provider is disabled")

type Provider interface {
	Name() string
	Send(ctx context.Context, device Device, message Message) error
}

type DisabledProvider struct{ ProviderName string }

func (p DisabledProvider) Name() string {
	if p.ProviderName == "" {
		return "disabled"
	}
	return p.ProviderName
}

func (DisabledProvider) Send(context.Context, Device, Message) error { return ErrProviderDisabled }
