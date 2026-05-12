module github.com/genvmoroz/lale-tg-client

go 1.26

require (
	github.com/genvmoroz/bot-engine v1.1.5
	github.com/genvmoroz/lale/service v1.0.0
	github.com/go-playground/validator/v10 v10.30.2
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/samber/lo v1.53.0
	github.com/sirupsen/logrus v1.9.4
	golang.org/x/text v0.37.0
	google.golang.org/grpc v1.81.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.51.0 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260511170946-3700d4141b60 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/genvmoroz/lale/service => ../service
