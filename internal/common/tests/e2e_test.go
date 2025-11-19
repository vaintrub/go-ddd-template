package tests

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vaintrub/go-ddd-template/internal/common/client"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"github.com/vaintrub/go-ddd-template/internal/common/genproto/users"
)

func TestCreateTraining(t *testing.T) {
	if os.Getenv("RUN_E2E_TESTS") != "1" {
		t.Skip("RUN_E2E_TESTS not set; skipping e2e flow")
	}

	t.Setenv("APP_ENV", "development")
	t.Setenv("ENV", "development")
	t.Setenv("ENV_NAME", "development")

	hour := RelativeDate(12, 12)

	userID := "8b112ac8-e1d5-4f91-ad9e-b30bc1495db6"
	trainerJWT := FakeTrainerJWT(t, uuid.New().String())
	attendeeJWT := FakeAttendeeJWT(t, userID)
	trainerHTTPClient := NewTrainerHTTPClient(t, trainerJWT)
	trainingsHTTPClient := NewTrainingsHTTPClient(t, attendeeJWT)
	usersHTTPClient := NewUsersHTTPClient(t, attendeeJWT)

	cfg := config.MustLoad(context.Background())

	usersGrpcClient, _, err := client.NewUsersClient(cfg.GRPC)
	require.NoError(t, err)

	// Cancel the training if exists and make the hour available
	trainings := trainingsHTTPClient.GetTrainings(t)
	for _, training := range trainings.Trainings {
		if training.Time.Equal(hour) {
			trainingsTrainerHTTPClient := NewTrainingsHTTPClient(t, trainerJWT)
			trainingsTrainerHTTPClient.CancelTraining(t, training.Uuid.String(), 200)
			break
		}
	}
	hours := trainerHTTPClient.GetTrainerAvailableHours(t, hour, hour)
	if len(hours) > 0 {
		for _, h := range hours[0].Hours {
			if h.Hour.Equal(hour) {
				trainerHTTPClient.MakeHourUnavailable(t, hour)
				break
			}
		}
	}

	trainerHTTPClient.MakeHourAvailable(t, hour)

	user := usersHTTPClient.GetCurrentUser(t)
	originalBalance := user.Balance

	_, err = usersGrpcClient.UpdateTrainingBalance(context.Background(), &users.UpdateTrainingBalanceRequest{
		UserId:       userID,
		AmountChange: 1,
	})
	require.NoError(t, err)

	user = usersHTTPClient.GetCurrentUser(t)
	require.Equal(t, originalBalance+1, user.Balance, "Attendee's balance should be updated")

	trainingUUID := trainingsHTTPClient.CreateTraining(t, "some note", hour)

	trainingsResponse := trainingsHTTPClient.GetTrainings(t)
	require.Len(t, trainingsResponse.Trainings, 1)
	require.Equal(t, trainingUUID, trainingsResponse.Trainings[0].Uuid.String(), "Attendee should see the training")

	user = usersHTTPClient.GetCurrentUser(t)
	require.Equal(t, originalBalance, user.Balance, "Attendee's balance should be updated after a training is scheduled")
}
