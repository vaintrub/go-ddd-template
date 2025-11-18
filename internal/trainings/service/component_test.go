package service

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vaintrub/go-ddd-template/internal/common/server"
	"github.com/vaintrub/go-ddd-template/internal/common/tests"
	"github.com/vaintrub/go-ddd-template/internal/trainings/ports"
)

func TestCreateTraining(t *testing.T) {
	t.Parallel()

	token := tests.FakeAttendeeJWT(t, uuid.New().String())
	client := tests.NewTrainingsHTTPClient(t, token)

	hour := tests.RelativeDate(10, 12)
	trainingUUID := client.CreateTraining(t, "some note", hour)

	trainingsResponse := client.GetTrainings(t)

	var trainingsUUIDs []string
	for _, t := range trainingsResponse.Trainings {
		trainingsUUIDs = append(trainingsUUIDs, t.Uuid.String())
	}

	require.Contains(t, trainingsUUIDs, trainingUUID)
}

func TestCancelTraining(t *testing.T) {
	t.Parallel()

	token := tests.FakeAttendeeJWT(t, uuid.New().String())
	client := tests.NewTrainingsHTTPClient(t, token)

	hour := tests.RelativeDate(10, 13)
	trainingUUID := client.CreateTraining(t, "some note", hour)

	client.CancelTraining(t, trainingUUID, http.StatusOK)

	trainingsResponse := client.GetTrainings(t)

	var trainingsUUIDs []string
	for _, t := range trainingsResponse.Trainings {
		trainingsUUIDs = append(trainingsUUIDs, t.Uuid.String())
	}

	require.NotContains(t, trainingsUUIDs, trainingUUID)
}

func startService(t *testing.T) bool {
	app := NewComponentTestApplication(context.Background())

	trainingsHTTPAddr := os.Getenv("TRAININGS_HTTP_ADDR")
	go server.RunHTTPServerOnAddr(trainingsHTTPAddr, func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpServer(app), router)
	})

	ok := tests.WaitForPort(trainingsHTTPAddr)
	if !ok {
		t.Log("Timed out waiting for trainings HTTP to come up")
	}

	return ok
}

func TestMain(m *testing.M) {
	dummyT := &testing.T{}
	if !startService(dummyT) {
		os.Exit(1)
	}

	os.Exit(m.Run())
}
