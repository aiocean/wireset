// Package discord_notify sends notification to discord
package discord_notify

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/aiocean/wireset/feature/discord-notify/event"
	"github.com/google/wire"
)

const (
	// FeatureName is the name of the feature.
	FeatureName = "discord-notify"
)

// DefaultWireset init dependencies
var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureNotify), "*"),
	event.NewNotifyDiscordOnInstallHandler,
)

// FeatureNotify struct
type FeatureNotify struct {
	EvtProcessor                  *cqrs.EventProcessor
	NotifyDiscordOnInstallHandler *event.NotifyDiscordOnInstallHandler
}

// Init init feature
func (f *FeatureNotify) Init() error {
	err := f.EvtProcessor.AddHandlers(f.NotifyDiscordOnInstallHandler)
	if err != nil {
		return err
	}
	return nil
}

// Name of the feature
func (f *FeatureNotify) Name() string {
	return FeatureName
}
