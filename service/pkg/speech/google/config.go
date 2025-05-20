package google

type Config struct {
	ProjectKeyJSON string `envconfig:"APP_GOOGLE_PROJECT_KEY_JSON"`
	StubEnabled    bool   `envconfig:"APP_GOOGLE_STUB_ENABLED" default:"false"`
}
