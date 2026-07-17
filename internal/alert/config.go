package alert

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultInterval = 30 * time.Second
	defaultCooldown = 15 * time.Minute
)

// Config holds Telegram alert settings.
type Config struct {
	NodeName  string
	Interval  time.Duration
	Cooldown  time.Duration
	Token     string
	ChatID    string
	DryRun    bool
}

// LoadConfig reads alert settings from flags and environment.
func LoadConfig(nodeName string, interval time.Duration, dryRun bool) Config {
	cfg := Config{
		NodeName: strings.TrimSpace(nodeName),
		Interval: interval,
		Cooldown: defaultCooldown,
		DryRun:   dryRun,
		Token:    strings.TrimSpace(os.Getenv("PMTOP_TELEGRAM_TOKEN")),
		ChatID:   strings.TrimSpace(os.Getenv("PMTOP_TELEGRAM_CHAT_ID")),
	}
	if cfg.Interval <= 0 {
		cfg.Interval = defaultInterval
	}
	if v := strings.TrimSpace(os.Getenv("PMTOP_NODE_NAME")); v != "" && cfg.NodeName == "" {
		cfg.NodeName = v
	}
	if v := envDuration("PMTOP_ALERT_INTERVAL"); v > 0 {
		cfg.Interval = v
	}
	if v := envDuration("PMTOP_ALERT_COOLDOWN"); v > 0 {
		cfg.Cooldown = v
	}
	return cfg
}

func (c Config) Enabled() bool {
	return c.Token != "" && c.ChatID != ""
}

func envDuration(key string) time.Duration {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return 0
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return 0
}
