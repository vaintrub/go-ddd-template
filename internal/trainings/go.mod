module github.com/vaintrub/go-ddd-template/internal/trainings

go 1.25.4

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/render v1.0.3
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/oapi-codegen/runtime v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.11.1
	github.com/vaintrub/go-ddd-template/internal/common v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.10
)

require (
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/go-chi/cors v1.2.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101 // indirect
	google.golang.org/grpc v1.76.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/vaintrub/go-ddd-template/internal/common => ../common
