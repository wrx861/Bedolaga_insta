package main

// ════════════════════════════════════════════════════════════════
// CONFIG
// ════════════════════════════════════════════════════════════════

type Config struct {
	InstallDir            string
	PanelInstalledLocally bool
	PanelDir              string
	DockerNetwork         string

	BotToken       string
	AdminIDs       string
	SupportUsername string

	RemnawaveAPIURL    string
	RemnawaveAPIKey    string
	RemnawaveAuthType  string
	RemnawaveUsername   string
	RemnawavePassword  string
	RemnawaveSecretKey string

	WebhookDomain string
	MiniappDomain string

	AdminNotificationsChatID string

	PostgresPassword    string
	KeepExistingVolumes bool
	OldPostgresPassword string

	WebhookSecretToken string
	WebAPIDefaultToken string
	BotRunMode         string
	WebhookURL         string
	WebAPIEnabled      string

	ReverseProxyType string
	SSLEmail         string
}
