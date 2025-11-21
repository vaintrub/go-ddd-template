package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	trainerHTTP "github.com/vaintrub/go-ddd-template/internal/common/client/trainer"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"github.com/vaintrub/go-ddd-template/internal/common/genproto/trainer"
	"github.com/vaintrub/go-ddd-template/internal/common/logs"
	"github.com/vaintrub/go-ddd-template/internal/common/server"
	"github.com/vaintrub/go-ddd-template/internal/common/tests"
	"github.com/vaintrub/go-ddd-template/internal/trainer/ports"
	"google.golang.org/grpc"
)

func TestHoursAvailability(t *testing.T) {
	t.Parallel()

	token := tests.FakeTrainerJWT(t, uuid.New().String())
	client := tests.NewTrainerHTTPClient(t, token)

	hour := tests.RelativeDate(11, 12)

	date := hour.Truncate(24 * time.Hour)
	from := date.AddDate(0, 0, -1)
	to := date.AddDate(0, 0, 1)

	getHours := func() []trainerHTTP.Hour {
		dates := client.GetTrainerAvailableHours(t, from, to)
		for _, d := range dates {
			if d.Date.Equal(date) {
				return d.Hours
			}
		}
		t.Fatalf("Date not found in dates: %+v", dates)
		return nil
	}

	findHour := func(hours []trainerHTTP.Hour, targetHour time.Time) *trainerHTTP.Hour {
		for _, h := range hours {
			// Compare using UTC to avoid timezone issues
			if h.Hour.UTC().Equal(targetHour.UTC()) {
				return &h
			}
		}
		return nil
	}

	client.MakeHourUnavailable(t, hour)
	foundHour := findHour(getHours(), hour)
	require.NotNil(t, foundHour, "hour should exist after making unavailable")
	require.False(t, foundHour.Available, "hour should not be available")

	code := client.MakeHourAvailable(t, hour)
	require.Equal(t, http.StatusNoContent, code)
	foundHour = findHour(getHours(), hour)
	require.NotNil(t, foundHour, "hour should exist after making available")
	require.True(t, foundHour.Available, "hour should be available")

	client.MakeHourUnavailable(t, hour)
	foundHour = findHour(getHours(), hour)
	require.NotNil(t, foundHour, "hour should exist after making unavailable again")
	require.False(t, foundHour.Available, "hour should not be available")
}

func TestUnauthorizedForAttendee(t *testing.T) {
	t.Parallel()

	token := tests.FakeAttendeeJWT(t, uuid.New().String())
	client := tests.NewTrainerHTTPClient(t, token)

	hour := tests.RelativeDate(11, 13)

	code := client.MakeHourAvailable(t, hour)
	require.Equal(t, http.StatusUnauthorized, code)
}

func startService(t *testing.T) bool {
	app := NewApplication(context.Background(), componentTestConfig())
	logger := logs.Init(config.LoggingConfig{Level: "INFO"})
	serverCfg := config.ServerConfig{MockAuth: true}

	trainerHTTPAddr := os.Getenv("TRAINER_HTTP_ADDR")
	go server.RunHTTPServerOnAddr(serverCfg, trainerHTTPAddr, logger, func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpServer(app), router)
	})

	trainerGrpcAddr := os.Getenv("TRAINER_GRPC_ADDR")
	go server.RunGRPCServerOnAddr(trainerGrpcAddr, logger, func(s *grpc.Server) {
		svc := ports.NewGrpcServer(app)
		trainer.RegisterTrainerServiceServer(s, svc)
	})

	ok := tests.WaitForPort(trainerHTTPAddr)
	if !ok {
		t.Log("Timed out waiting for trainer HTTP to come up")
		return false
	}

	ok = tests.WaitForPort(trainerGrpcAddr)
	if !ok {
		t.Log("Timed out waiting for trainer gRPC to come up")
	}

	return ok
}

func TestMain(m *testing.M) {
	// Create a dummy testing.T for logging in setup
	// In TestMain we can't use t.Log, so we use fmt.Fprintf to stderr
	dummyT := &testing.T{}
	if !startService(dummyT) {
		fmt.Fprintln(os.Stderr, "Timed out waiting for services to come up")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func componentTestConfig() config.Config {
	cfg := config.DefaultConfig()
	cfg.Database.URL = os.Getenv("DATABASE_URL")
	cfg.GRPC.TrainerAddr = os.Getenv("TRAINER_GRPC_ADDR")
	cfg.GRPC.UsersAddr = os.Getenv("USERS_GRPC_ADDR")
	cfg.Server.Port = 0
	return cfg
}
