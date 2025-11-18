module github.com/vaintrub/go-ddd-template/tools/c4

go 1.25.4

require (
	github.com/krzysztofreczek/go-structurizr v0.1.55
	github.com/vaintrub/go-ddd-template/internal/trainer v0.0.0
	github.com/vaintrub/go-ddd-template/internal/trainings v0.0.0
)

require (
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/cnf/structhash v0.0.0-20250313080605-df4c6cc74a9a // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/go-chi/chi/v5 v5.2.3 // indirect
	github.com/go-chi/render v1.0.3 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.6 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/vaintrub/go-ddd-template/internal/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101 // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/vaintrub/go-ddd-template/internal/common => ../../internal/common
	github.com/vaintrub/go-ddd-template/internal/trainer => ../../internal/trainer
	github.com/vaintrub/go-ddd-template/internal/trainings => ../../internal/trainings
)
