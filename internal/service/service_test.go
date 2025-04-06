package service

import (
	"errors"
	"testing"

	"github.com/lovelaze/nebula-sync/internal/config"
	syncmock "github.com/lovelaze/nebula-sync/internal/mocks/sync"
	webhookmock "github.com/lovelaze/nebula-sync/internal/mocks/webhook"
	"github.com/lovelaze/nebula-sync/internal/pihole/model"
	"github.com/stretchr/testify/require"
)

func TestRun_full(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: true,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	webhook := webhookmock.NewWebhookClient(t)
	target.On("FullSync", conf.Sync).Return(nil)
	webhook.On("Success").Return(nil)

	service := Service{
		target:  target,
		conf:    conf,
		webhook: webhook,
	}

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "FullSync", conf.Sync)
}

func TestRun_selective(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: false,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	webhook := webhookmock.NewWebhookClient(t)
	target.On("SelectiveSync", conf.Sync).Return(nil)
	webhook.On("Success").Return(nil)

	service := Service{
		target:  target,
		conf:    conf,
		webhook: webhook,
	}

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "SelectiveSync", conf.Sync)
}

func TestRun_webhook_success(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: false,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	webhook := webhookmock.NewWebhookClient(t)

	target.On("SelectiveSync", conf.Sync).Return(nil)
	webhook.On("Success").Return(nil)

	service := Service{
		target:  target,
		conf:    conf,
		webhook: webhook,
	}

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "SelectiveSync", conf.Sync)
	webhook.AssertCalled(t, "Success")
	webhook.AssertNotCalled(t, "Failure")
}

func TestRun_webhook_failure(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: false,
			Cron:     nil,
		},
	}

	syncErr := errors.New("sync failed")
	target := syncmock.NewTarget(t)
	webhook := webhookmock.NewWebhookClient(t)

	target.On("SelectiveSync", conf.Sync).Return(syncErr)
	webhook.On("Failure").Return(nil)

	service := Service{
		target:  target,
		conf:    conf,
		webhook: webhook,
	}

	err := service.Run()
	require.ErrorIs(t, err, syncErr)

	target.AssertCalled(t, "SelectiveSync", conf.Sync)
	webhook.AssertCalled(t, "Failure")
	webhook.AssertNotCalled(t, "Success")
}

func TestRun_webhook_error_does_not_affect_result(t *testing.T) {
	conf := config.Config{
		Primary:  model.PiHole{},
		Replicas: []model.PiHole{},
		Sync: &config.Sync{
			FullSync: true,
			Cron:     nil,
		},
	}

	target := syncmock.NewTarget(t)
	webhook := webhookmock.NewWebhookClient(t)

	target.On("FullSync", conf.Sync).Return(nil)
	webhook.On("Success").Return(errors.New("webhook failed"))

	service := Service{
		target:  target,
		conf:    conf,
		webhook: webhook,
	}

	err := service.Run()
	require.NoError(t, err)

	target.AssertCalled(t, "FullSync", conf.Sync)
	webhook.AssertCalled(t, "Success")
}
