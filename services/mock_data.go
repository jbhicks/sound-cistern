package services

import (
	"strings"

	"github.com/jbhicks/sound-cistern/views"
	"github.com/jbhicks/sound-cistern/views/components"
)

type MockDataService struct{}

func NewMockDataService() *MockDataService {
	return &MockDataService{}
}

type MockTrack struct {
	ID       string
	Title    string
	Artist   string
	Genre    string
	Duration int64
	Plays    int64
	Likes    int64
	Reposts  int64
}

func (m *MockDataService) GetStreamTracks(durationMin int, searchQuery string) string {
	mockTracks := []MockTrack{
		{"1", "Test Track One", "Artist A", "Electronic", 180000, 1500, 250, 50},
		{"2", "Another Track", "Artist B", "House", 240000, 3200, 480, 120},
		{"3", "Short Track", "Artist C", "Techno", 60000, 890, 120, 25},
	}

	return m.renderTracks(mockTracks, durationMin, searchQuery)
}

func (m *MockDataService) GetFavoritesTracks(durationMin int, searchQuery string) string {
	mockFavorites := []MockTrack{
		{"101", "Favorite Track One", "Artist X", "Electronic", 180000, 1500, 250, 50},
		{"102", "Another Favorite", "Artist Y", "House", 240000, 3200, 480, 120},
		{"103", "Liked Song", "Artist Z", "Techno", 60000, 890, 120, 25},
	}

	return m.renderTracks(mockFavorites, durationMin, searchQuery)
}

func (m *MockDataService) renderTracks(tracks []MockTrack, durationMin int, searchQuery string) string {
	var sb strings.Builder
	sb.WriteString("<div class=\"stream-flip-grid\">")

	searchLower := strings.ToLower(searchQuery)

	for _, t := range tracks {
		// Filter by duration
		durationMinutes := t.Duration / 60000
		if durationMin > 0 && durationMinutes < int64(durationMin) {
			continue
		}

		// Filter by search
		if searchQuery != "" {
			titleMatch := strings.Contains(strings.ToLower(t.Title), searchLower)
			artistMatch := strings.Contains(strings.ToLower(t.Artist), searchLower)
			if !titleMatch && !artistMatch {
				continue
			}
		}

		artworkURL := "https://picsum.photos/seed/" + t.ID + "/500/500"
		components.StreamFlipCard(
			t.ID,
			t.Title,
			t.Artist,
			t.Genre,
			t.Duration,
			artworkURL,
			t.Plays,
			t.Likes,
			t.Reposts,
		).Render(nil, &sb)
	}

	sb.WriteString("</div>")
	return sb.String()
}

func (m *MockDataService) GetTracksForTemplate() []views.Track {
	return []views.Track{
		{
			TrackID:          "1",
			TrackTitle:       "Test Track One",
			ArtistName:       "Artist A",
			Genre:            "Electronic",
			TrackDuration:    180000,
			ArtworkURL:       "https://picsum.photos/seed/1/500/500",
			PlaybackCount:    1500,
			FavoritingsCount: 250,
			RepostsCount:     50,
		},
		{
			TrackID:          "2",
			TrackTitle:       "Another Track",
			ArtistName:       "Artist B",
			Genre:            "House",
			TrackDuration:    240000,
			ArtworkURL:       "https://picsum.photos/seed/2/500/500",
			PlaybackCount:    3200,
			FavoritingsCount: 480,
			RepostsCount:     120,
		},
		{
			TrackID:          "3",
			TrackTitle:       "Short Track",
			ArtistName:       "Artist C",
			Genre:            "Techno",
			TrackDuration:    60000,
			ArtworkURL:       "https://picsum.photos/seed/3/500/500",
			PlaybackCount:    890,
			FavoritingsCount: 120,
			RepostsCount:     25,
		},
	}
}

func (m *MockDataService) GetFavoritesForTemplate() []views.Track {
	return []views.Track{
		{
			TrackID:          "101",
			TrackTitle:       "Favorite Track One",
			ArtistName:       "Artist X",
			Genre:            "Electronic",
			TrackDuration:    180000,
			ArtworkURL:       "https://picsum.photos/seed/101/500/500",
			PlaybackCount:    1500,
			FavoritingsCount: 250,
			RepostsCount:     50,
		},
		{
			TrackID:          "102",
			TrackTitle:       "Another Favorite",
			ArtistName:       "Artist Y",
			Genre:            "House",
			TrackDuration:    240000,
			ArtworkURL:       "https://picsum.photos/seed/102/500/500",
			PlaybackCount:    3200,
			FavoritingsCount: 480,
			RepostsCount:     120,
		},
		{
			TrackID:          "103",
			TrackTitle:       "Liked Song",
			ArtistName:       "Artist Z",
			Genre:            "Techno",
			TrackDuration:    60000,
			ArtworkURL:       "https://picsum.photos/seed/103/500/500",
			PlaybackCount:    890,
			FavoritingsCount: 120,
			RepostsCount:     25,
		},
	}
}
