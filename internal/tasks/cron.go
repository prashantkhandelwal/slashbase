package tasks

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"slashbase.com/backend/internal/config"
)

func InitCron() {

	if !config.IsLive() {
		return
	}

	scheduler := gocron.NewScheduler(time.UTC)

	telemetryPings(scheduler)

	scheduler.StartAsync()
}

func telemetryPings(s *gocron.Scheduler) {
	s.Every(1).Day().Do(func() {
		telemetryID := config.GetTelemetryID()
		if telemetryID == "TEST" {
			return
		}
		values := map[string]interface{}{
			"api_key": "phc_XSWvMvnTUEH9pLJDVmYfaKaKH8QZtK5fJO8NIiFoNwv",
			"event":   "Telemetry Ping",
			"properties": map[string]string{
				"distinct_id": telemetryID,
				"version":     config.VERSION,
			},
		}
		json_data, err := json.Marshal(values)
		if err != nil {
			log.Fatal(err)
		}

		http.Post("https://app.posthog.com/capture/", "application/json",
			bytes.NewBuffer(json_data))
	})
}