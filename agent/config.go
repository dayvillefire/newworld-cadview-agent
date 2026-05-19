package agent

import "os"

// LoadConfigFromEnv overrides Agent fields with environment variable values
// when those variables are set. Values from env take precedence over the
// fields already set on the Agent struct.
func (a *Agent) LoadConfigFromEnv() {
	if v := os.Getenv(ENV_BASE_URL); v != "" {
		a.BaseUrl = v
	}
	if v := os.Getenv(ENV_USERNAME); v != "" {
		a.Username = v
	}
	if v := os.Getenv(ENV_PASSWORD); v != "" {
		a.Password = v
	}
	if v := os.Getenv(ENV_FDID); v != "" {
		a.FDID = v
	}
}
