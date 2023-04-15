package google

type Config struct {
	ProjectKeyFile string `envconfig:"APP_GOOGLE_PROJECT_KEY_FILE" required:"true" json:"ProjectKeyFile,omitempty"`
}
