package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"tubr/config"
	"tubr/models"
	"tubr/store"
)

var port = ":7070"

func schedule(w http.ResponseWriter, r *http.Request) {
	schedules, err := db.GetSchedules(r.Context(), &store.ScheduleQueryArgs{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	newVideos := []models.Video{}

	// Determine what new videos must be created
	for _, schedule := range schedules {
		count, err := db.CountVideosByScheduleID(r.Context(), schedule.ID)
		if err != nil {
			log.Printf("failed to count videos by schedule ID %+v", err)
			continue
		}

		cursor := schedule.StartDate
		var calculatedCount int
		for cursor.Before(time.Now()) {
			calculatedCount++
			cursor = cursor.Add((time.Hour * 24) * time.Duration(schedule.FrequencyDays))
		}

		log.Printf("videos matching schedule ID: %d, calculated count: %d\n", count, calculatedCount)

		if count < calculatedCount {
			log.Println("generating new videos")
			videos, err := db.GetVideosByScheduleId(r.Context(), schedule.ID)
			if err != nil {
				log.Printf("failed to get videos by schedule ID: %+v", err)
				continue
			}
			log.Printf("got %d videos from the database\n", len(videos))

			var lastVid models.Video
			if len(videos) >= 1 {
				lastVid = videos[0]
				log.Printf("got last vid of sequence: %+v\n", lastVid)
			} else {
				startDate := schedule.StartDate.Add(-time.Second)
				log.Printf("got start date at %s\n", startDate.Format(time.RFC3339))

				lastVid.StartDate = &startDate
				lastVid.EndDate = &startDate
			}

			for lastVid.EndDate.Before(time.Now()) {
				t1 := lastVid.EndDate.Add(time.Second)
				t2 := t1.Add(time.Hour*24*time.Duration(schedule.FrequencyDays) - time.Second)
				releaseDate := t2.Add(12 * time.Hour)

				newVid := models.Video{
					StartDate:        &t1,
					EndDate:          &t2,
					ScheduleID:       schedule.ID,
					CacheDestination: nil,
					Uploaded:         false,
					ReleaseDate:      &releaseDate,
				}

				newVideos = append(newVideos, newVid)
				lastVid = newVid
			}
		}
	}

	if err = db.InsertVideos(r.Context(), newVideos); err != nil {
		log.Printf("failed to insert new videos into database, %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if err := upcoming(r.Context(), newVideos, schedules); err != nil {
		log.Printf("failed to create upcoming videos, %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("scheduled new videos"))
}

var (
	db *store.DB
)

func upcoming(ctx context.Context, newVideos []models.Video, schedules []models.Schedule) error {

	schedulesByScheduleID := map[int]models.Schedule{}
	for _, schedule := range schedules {
		schedulesByScheduleID[schedule.ID] = schedule
	}

	for _, newVideo := range newVideos {
		// Check if the release date is less than 2 days from now
		// and the end date is in the past
		if newVideo.ReleaseDate.Before(time.Now().Add(48*time.Hour)) &&
			newVideo.EndDate.Before(time.Now()) {

			// If the cache destination is not set, create the video
			if newVideo.CacheDestination == nil {
				if schedulesByScheduleID[newVideo.ScheduleID].GameID == "" {
					log.Printf("GameID is empty for Schedule ID: %d", newVideo.ScheduleID)
					continue // Skip to the next video
				}

				startDateStr := newVideo.StartDate.Format("2006-01-02") // Use the Go time layout for YYYY-MM-DD
				endDateStr := newVideo.EndDate.Format("2006-01-02")     // Go time layout for formatting

				url := fmt.Sprintf("http://localhost:8080/processHandler?t1=%s&t2=%s&game_id=%s&target_duration=%d",
					startDateStr,
					endDateStr,
					schedulesByScheduleID[newVideo.ScheduleID].GameID,
					schedulesByScheduleID[newVideo.ScheduleID].TargetDurationSeconds)

				// Log the outgoing request to localhost:8080
				log.Printf("Sending POST request to: %s", url)

				req, err := http.NewRequest("POST", url, nil)

				if err != nil {
					log.Fatalf("Error creating request: %v", err) // Log and stop execution if an error occurred
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("Error sending request: %v", err)
					continue // Move to the next video
				}

				// Read the body before decoding
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Error reading response body: %v", err)
					continue // Move to the next video

				}

				// Handle "OK" response
				if string(body) == "OK" {
					log.Println("Received OK response, moving to next video.")
					continue // Skip to the next video in the loop
				}

				processIdStr := string(body[1 : len(body)-2])
				log.Printf("response body (trimmed): %s", processIdStr)
				cacheDestination := "C:/Users/will/tubr/" + processIdStr + "/compilation.mp4"

				// Update the CacheDestination field in the database
				err = db.UpdateCacheDestination(ctx, cacheDestination, *newVideo.StartDate, newVideo.ScheduleID)
				if err != nil {
					log.Printf("Failed to update cache destination in the database: %v", err)
					continue // Skip to the next video
				}

				log.Printf("Updated cache destination to: %s", cacheDestination)

				newVideo.CacheDestination = &cacheDestination

				defer resp.Body.Close()

				//     // Create a new HTTP request to upload the video
				//     req, err := http.NewRequest("POST", "/upload", nil)
				//     if err != nil {
				//         return err  // Handle request creation error
				//     }

				//     // Execute the request (this part would include handling the request)
				//     // executeRequest(req)
				// } else {
				//     // If cache destination is already set, just upload the video
				//     req, err := http.NewRequest("POST", "/upload", nil)
				//     if err != nil {
				//         return err  // Handle request creation error
				//     }

				//     // Execute the request (this part would include handling the request)
				//     // executeRequest(req)
			}
		}
	}

	return nil
}

func main() {
	cfg := config.FromENV()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt)

	ctx := context.Background()

	var err error
	db, err = store.Open(ctx, cfg.PostgresConfig)
	if err != nil {
		log.Printf("error: failed to connect to database %+v", err)
		return
	}

	_, cancel := context.WithCancel(ctx)

	srv := &http.Server{Addr: port}

	go func() {
		<-kill
		log.Println("exiting gracefully")
		cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("error: failed to exit server gracefully %s\n", err)
		}
	}()

	http.HandleFunc("/schedule", schedule)

	log.Printf("Starting server on port %s\n", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Println(err)
		return
	}
}
