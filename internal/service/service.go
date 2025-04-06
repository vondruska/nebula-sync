package service

import (
	"fmt"
	"net/http"

	"github.com/lovelaze/nebula-sync/internal/sync/retry"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/internal/pihole"
	"github.com/lovelaze/nebula-sync/internal/sync"
	"github.com/lovelaze/nebula-sync/version"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

type Service struct {
	target sync.Target
	conf   config.Config
}

func Init() (*Service, error) {
	conf := config.Config{}
	if err := conf.Load(); err != nil {
		return nil, err
	}

	httpClient := conf.Client.NewHttpClient()
	retry.Init(conf.Client)

	primary := pihole.NewClient(conf.Primary, httpClient)
	var replicas []pihole.Client
	for _, replica := range conf.Replicas {
		replicas = append(replicas, pihole.NewClient(replica, httpClient))
	}

	return &Service{
		target: sync.NewTarget(primary, replicas),
		conf:   conf,
	}, nil
}

func (service *Service) Run() error {
	log.Info().Msgf("Starting nebula-sync %s", version.Version)
	log.Debug().Str("config", service.conf.String()).Msgf("Settings")

	if err := service.doSync(service.target); err != nil {
		return err
	}

	if service.conf.Sync.Cron != nil {
		return service.startCron(func() {
			if err := service.doSync(service.target); err != nil {
				log.Error().Err(err).Msg("Sync failed")
			}
		})
	}

	return nil
}

func (service *Service) doSync(t sync.Target) (err error) {
	if service.conf.Sync.FullSync {
		err = t.FullSync(service.conf.Sync)
	} else {
		err = t.SelectiveSync(service.conf.Sync)
	}

	if err != nil {
		if service.conf.Sync.FailureWebhookURL != "" {
			if err := sendWebhook(service.conf.Sync.FailureWebhookURL); err != nil {
				log.Error().Err(err).Msg("Failed to send failure webhook")
			}
		}
	} else {
		log.Info().Msg("Sync completed")

		if service.conf.Sync.SuccessWebhookURL != "" {
			if err := sendWebhook(service.conf.Sync.SuccessWebhookURL); err != nil {
				log.Error().Err(err).Msg("Failed to send success webhook")
			}
		}
	}

	return err
}

func (service *Service) startCron(cmd func()) error {
	cron := cron.New()

	if _, err := cron.AddFunc(*service.conf.Sync.Cron, cmd); err != nil {
		return fmt.Errorf("cron job: %w", err)
	}

	cron.Run()
	return nil
}

func sendWebhook(webhookUrl string) error {
	req, err := http.NewRequest("POST", webhookUrl, nil)
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("nebula-sync/%s", version.Version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
