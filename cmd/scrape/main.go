package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	vodmodule_stats "github.com/zsiec/vodmodule-stats"
)

type Config struct {
	Namespace      string        `default:"jigsaw-dev"`
	StatusPath     string        `default:"/status"`
	ScrapeInterval time.Duration `default:"60s"`
}

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("svc", "vodmodule-stats").Logger()

	cfg, err := parseConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("parsing config")
	}

	logger = logger.With().
		Str("namespace", cfg.Namespace).
		Str("status_path", cfg.StatusPath).
		Str("scrapeInterval", cfg.ScrapeInterval.String()).
		Logger()

	scraper := vodmodule_stats.PodScraper{
		Namespace:  cfg.Namespace,
		StatusPath: cfg.StatusPath,
		Logger:     logger,
	}

	go func() {
		for {
			if err := scraper.Scrape(); err != nil {
				logger.Err(err).Msg("scrape error")
			}
			time.Sleep(cfg.ScrapeInterval)
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	s := <-c
	close(c)

	logger.Info().Msgf("caught termination signal %s", s)
}

func parseConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("vs", &cfg)

	return cfg, err
}
