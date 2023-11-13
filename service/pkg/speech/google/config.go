package google

type Config struct {
	ProjectKeyJSON string `envconfig:"APP_GOOGLE_PROJECT_KEY_JSON" required:"true" json:"ProjectKeyJSON,omitempty"`
}
