package StatsIO

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sa-kemper/peertubestats/pkg/peertubeApi"
)

func Test_mergeVideoDB(t *testing.T) {
	type lenStruct struct {
		currentDb int
		inputDb   int
		deletedDb int
	}

	type args struct {
		currentData   *sync.Map
		inputDatabase *sync.Map
		deletedDb     *sync.Map
		recordedTs    time.Time
	}
	testVideo := peertubeApi.VideoData{
		ID:                    1,
		UUID:                  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		ShortUUID:             "0000000000001",
		IsLive:                false,
		LiveSchedules:         nil,
		CreatedAt:             "today",
		PublishedAt:           "today",
		UpdatedAt:             "today",
		OriginallyPublishedAt: "today",
		Category: peertubeApi.Metadata{
			ID:    "nil",
			Label: "aaa",
		},
		TruncatedDescription: "A description",
		Duration:             69,
		AspectRatio:          16 / 9,
		IsLocal:              true,
		Name:                 "Successfull video",
		ThumbnailPath:        "/lazy-static/thumbnails/0d0022c4-5182-4d8f-9cd4-5d6d7ecf7f17.jpg",
		PreviewPath:          "/lazy-static/thumbnails/0d0022c4-5182-4d8f-9cd4-5d6d7ecf7f17.jpg",
		EmbedPath:            "/lazy-static/thumbnails/0d0022c4-5182-4d8f-9cd4-5d6d7ecf7f17.jpg",
		Views:                1,
		Likes:                1,
		Dislikes:             1,
		Comments:             1,
		State:                peertubeApi.Metadata{},
		ScheduledUpdate:      peertubeApi.ScheduledUpdate{},
		Blacklisted:          false,
		BlacklistedReason:    "",
		Account: peertubeApi.Account{
			ID:          1,
			Name:        "derAccount",
			DisplayName: "derAccount",
			URL:         "",
			Host:        "",
			Avatars: []peertubeApi.Avatar{
				{
					Path:      "https://placehold.co/600x400?text=derAccount",
					Width:     400,
					Height:    600,
					CreatedAt: "today",
					UpdatedAt: "today",
				},
			},
		},
		Channel: peertubeApi.Channel{
			ID:          1,
			Name:        "derAccount",
			DisplayName: "derAccount",
			URL:         "https://placehold.co/600x400?text=derAccount",
			Host:        "example.com",
			Avatars: []peertubeApi.Avatar{
				{
					Path:      "https://placehold.co/600x400?text=derAccount",
					Width:     400,
					Height:    600,
					CreatedAt: "today",
					UpdatedAt: "today",
				},
			},
		},
	}
	dbWithVid := &sync.Map{}
	dbWithVid.Store(testVideo.ID, testVideo)
	tests := []struct {
		name            string
		args            args
		expectedLengths lenStruct
		wantErr         bool
	}{
		{
			name: "From empty to full",
			args: args{
				currentData:   &sync.Map{},
				inputDatabase: dbWithVid,
				deletedDb:     &sync.Map{},
			},
			expectedLengths: lenStruct{
				currentDb: 1,
				inputDb:   1,
				deletedDb: 0,
			},
			wantErr: false,
		},
		{
			name: "From full to empty",
			args: args{
				currentData:   dbWithVid,
				inputDatabase: &sync.Map{},
				deletedDb:     &sync.Map{},
			},
			expectedLengths: lenStruct{
				currentDb: 1,
				inputDb:   0,
				deletedDb: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mergeVideoDB(tt.args.currentData, tt.args.inputDatabase, tt.args.deletedDb, tt.args.recordedTs)
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeVideoDB() error = %v, wantErr %v", err, tt.wantErr)
			}
			var currentDbLen, inputDbLen, deletedDbLen int
			tt.args.currentData.Range(func(key, value interface{}) bool { currentDbLen++; return true })
			tt.args.inputDatabase.Range(func(key, value interface{}) bool { inputDbLen++; return true })
			tt.args.deletedDb.Range(func(key, value interface{}) bool { deletedDbLen++; return true })

			if !reflect.DeepEqual(tt.expectedLengths.inputDb, inputDbLen) {
				t.Errorf("expexted input length is not met. expected: %v, actual: %v", tt.expectedLengths.inputDb, inputDbLen)
			}
			if !reflect.DeepEqual(tt.expectedLengths.currentDb, currentDbLen) {
				t.Errorf("expexted currentDb length is not met. expected: %v, actual: %v", tt.expectedLengths.currentDb, currentDbLen)
			}
			if !reflect.DeepEqual(tt.expectedLengths.deletedDb, deletedDbLen) {
				t.Errorf("expexted deletedDb length is not met. expected: %v, actual: %v", tt.expectedLengths.deletedDb, deletedDbLen)
			}
		})
	}
}
