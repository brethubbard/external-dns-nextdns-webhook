package main

import (
	"os"
	"time"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
	"sigs.k8s.io/external-dns/provider/webhook"
)

type Config struct {
	DomainFilter                []string
	NextDNSAPIKey               string
	NextDNSProfileId            string
	LogFormat                   string
	DryRun                      bool
	LogLevel                    string
	WebhookProviderReadTimeout  time.Duration
	WebhookProviderWriteTimeout time.Duration
	WebhookProviderURL          string
}

type NextDnsWebhook struct {
	provider provider.Provider
}

var defaultConfig = Config{
	DomainFilter:                []string{},
	NextDNSAPIKey:               "",
	NextDNSProfileId:            "",
	LogFormat:                   "text",
	DryRun:                      false,
	LogLevel:                    log.InfoLevel.String(),
	WebhookProviderReadTimeout:  5 * time.Second,
	WebhookProviderWriteTimeout: 10 * time.Second,
	WebhookProviderURL:          "http://localhost:8888",
}

// Version is the current version of the app, generated at build time
var Version = "unknown"

func allLogLevelsAsStrings() []string {
	var levels []string
	for _, level := range log.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}

func (cfg *Config) ParseFlags(args []string) error {
	app := kingpin.New("nextdns-webhook", "A webhook for ExternalDNS to sync records with NextDNS.\n\nNote that all flags may be replaced with env vars - `--flag` -> `NEXTDNS_WEBHOOK_FLAG=1` or `--flag value` -> NEXTDNS_WEBHOOK_FLAG=value`")
	app.Version(Version)
	app.DefaultEnvars()

	app.Flag("domain-filter", "Limit possible target zones by a domain suffix; specify multiple times for multiple domains (optional)").Default("").StringsVar(&cfg.DomainFilter)
	app.Flag("api-key", "When using the NextDNS provider, specify the API key for your NextDNS account").Default(defaultConfig.NextDNSAPIKey).StringVar(&cfg.NextDNSAPIKey)
	app.Flag("profile-id", "When using the NextDNS provider, specify the profile id you want external dns to sync").Default(defaultConfig.NextDNSProfileId).StringVar(&cfg.NextDNSProfileId)
	app.Flag("log-format", "The format in which log messages are printed (default: text, options: text, json)").Default(defaultConfig.LogFormat).EnumVar(&cfg.LogFormat, "text", "json")
	app.Flag("dry-run", "When enabled, prints DNS record changes rather than actually performing them (default: disabled)").BoolVar(&cfg.DryRun)
	app.Flag("log-level", "Set the level of logging. (default: info, options: panic, debug, info, warning, error, fatal").Default(defaultConfig.LogLevel).EnumVar(&cfg.LogLevel, allLogLevelsAsStrings()...)
	app.Flag("read-timeout", "The read timeout for the webhook provider in duration format (default: 5s)").Default(defaultConfig.WebhookProviderReadTimeout.String()).DurationVar(&cfg.WebhookProviderReadTimeout)
	app.Flag("write-timeout", "The write timeout for the webhook provider in duration format (default: 10s)").Default(defaultConfig.WebhookProviderWriteTimeout.String()).DurationVar(&cfg.WebhookProviderWriteTimeout)

	_, err := app.Parse(args)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// instantiate the config
	cfg := Config{}
	if err := cfg.ParseFlags(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing error: %v", err)
	}

	if cfg.LogFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}
	if cfg.DryRun {
		log.Info("running in dry-run mode. No changes to DNS records will be made.")
	}

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to parse log level: %v", err)
	}
	log.SetLevel(logLevel)

	// instantiate the dns provider
	dnsProvider, err := NewNextDNSProvider(NextDNSConfig{
		DryRun:           cfg.DryRun,
		NextDNSAPIKey:    cfg.NextDNSAPIKey,
		NextDNSProfileId: cfg.NextDNSProfileId,
		DomainFilter:     endpoint.NewDomainFilter(cfg.DomainFilter),
	})

	if err != nil {
		log.Fatalf("failed to create NextDNS provider: %v", err)
	}

	webhook.StartHTTPApi(dnsProvider, make(chan struct{}), cfg.WebhookProviderReadTimeout, cfg.WebhookProviderWriteTimeout, "127.0.0.1:8888")

}
