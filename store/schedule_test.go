package store

import (
	"context"
	"testing"
	"time"
	"tubr/models"
)

func TestInsertSchedule(t *testing.T) {
	scheduleId := 123
	err := testDatabase.InsertSchedule(context.Background(), &models.Schedule{
		ID:                    scheduleId,
		Label:                 "aoiuwjepfaiowjef",
		GameID:                "OIWJED",
		Broadcaster:           "POWIJEF",
		Platform:              "poaiwjef",
		Language:              "we",
		BroadcasterID:         "OWIEJF",
		StartDate:             time.Time{},
		FrequencyDays:         3,
		ClipTimeMaxSeconds:    23,
		RepeatBroadcaster:     false,
		TargetDurationSeconds: 892,
		WebhookURL:            "http://oafiwpeofihawepofih",
	})

	if err != nil {
		t.Error(err)
	}
}
