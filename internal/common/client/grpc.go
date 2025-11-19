package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"github.com/vaintrub/go-ddd-template/internal/common/genproto/trainer"
	"github.com/vaintrub/go-ddd-template/internal/common/genproto/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewTrainerClient(cfg config.GRPCConfig) (client trainer.TrainerServiceClient, close func() error, err error) {
	return newClient(cfg.TrainerAddr, cfg, func(conn grpc.ClientConnInterface) trainer.TrainerServiceClient {
		return trainer.NewTrainerServiceClient(conn)
	})
}

func NewUsersClient(cfg config.GRPCConfig) (client users.UsersServiceClient, close func() error, err error) {
	return newClient(cfg.UsersAddr, cfg, func(conn grpc.ClientConnInterface) users.UsersServiceClient {
		return users.NewUsersServiceClient(conn)
	})
}

type clientFactory[T any] func(grpc.ClientConnInterface) T

func newClient[T any](addr string, cfg config.GRPCConfig, factory clientFactory[T]) (T, func() error, error) {
	var zero T
	if addr == "" {
		return zero, func() error { return nil }, fmt.Errorf("grpc address is required")
	}

	opts, err := grpcDialOpts(addr, cfg)
	if err != nil {
		return zero, func() error { return nil }, err
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return zero, func() error { return nil }, err
	}

	return factory(conn), conn.Close, nil
}

func grpcDialOpts(addr string, cfg config.GRPCConfig) ([]grpc.DialOption, error) {
	if cfg.NoTLS {
		return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, nil
	}

	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Wrap(err, "cannot load root CA cert")
	}

	tlsConfig := &tls.Config{
		RootCAs:    systemRoots,
		MinVersion: tls.VersionTLS12,
	}
	if cfg.CAFile != "" {
		certPool, loadErr := loadCertPool(cfg.CAFile)
		if loadErr != nil {
			return nil, loadErr
		}
		tlsConfig.RootCAs = certPool
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		grpc.WithPerRPCCredentials(newMetadataServerToken(addr)),
	}, nil
}

func loadCertPool(path string) (*x509.CertPool, error) {
	// #nosec G304 - CA file path is provided via configuration managed by trusted deployment.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read CA file")
	}
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(data); !ok {
		return nil, fmt.Errorf("failed to append CA certs from %s", path)
	}
	return pool, nil
}
