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

func (m *MockDataService) GetStreamTracks(durationMin int, searchQuery string, offset int, limit int) string {
	mockTracks := []MockTrack{
		{"1", "Test Track One", "Artist A", "Electronic", 180000, 1500, 250, 50},
		{"2", "Another Track", "Artist B", "House", 240000, 3200, 480, 120},
		{"3", "Short Track", "Artist C", "Techno", 60000, 890, 120, 25},
		{"4", "Long Track One", "Artist D", "Ambient", 780000, 500, 100, 20},
		{"5", "Long Track Two", "Artist E", "Ambient", 900000, 600, 150, 30},
		{"6", "Long Track Three", "Artist F", "Ambient", 840000, 450, 80, 15},
		{"7", "Long Track Four", "Artist G", "Ambient", 1200000, 800, 200, 50},
		{"8", "Long Track Five", "Artist H", "Ambient", 960000, 700, 180, 40},
		{"9", "Long Track Six", "Artist I", "Ambient", 1080000, 900, 250, 60},
		{"10", "Long Track Seven", "Artist J", "Ambient", 720000, 400, 70, 10},
		{"11", "Long Track Eight", "Artist K", "Ambient", 1140000, 850, 220, 55},
		{"12", "Long Track Nine", "Artist L", "Ambient", 660000, 350, 60, 8},
		{"13", "Long Track Ten", "Artist M", "Ambient", 1320000, 1000, 300, 75},
		{"14", "Medium-Long One", "Artist N", "Electronic", 480000, 1500, 200, 45},
		{"15", "Medium-Long Two", "Artist O", "House", 540000, 1800, 280, 60},
		{"16", "Medium-Long Three", "Artist P", "Techno", 510000, 1600, 240, 52},
		{"17", "Medium-Long Four", "Artist Q", "Ambient", 570000, 2000, 320, 70},
		{"18", "Medium-Long Five", "Artist R", "Electronic", 450000, 1400, 180, 38},
		{"19", "Track Nineteen", "Artist S", "House", 300000, 2500, 350, 80},
		{"20", "Track Twenty", "Artist T", "Electronic", 320000, 2800, 400, 95},
		{"21", "Track Twenty-One", "Artist U", "Ambient", 380000, 3000, 450, 110},
		{"22", "Track Twenty-Two", "Artist V", "Techno", 350000, 2700, 380, 90},
		{"23", "Track Twenty-Three", "Artist W", "House", 280000, 2200, 300, 70},
		{"24", "Track Twenty-Four", "Artist X", "Electronic", 400000, 3100, 480, 120},
		{"25", "Track Twenty-Five", "Artist Y", "Ambient", 420000, 3400, 520, 130},
		{"26", "Track Twenty-Six", "Artist Z", "Techno", 360000, 2900, 420, 105},
		{"27", "Track Twenty-Seven", "Artist AA", "House", 310000, 2400, 340, 85},
		{"28", "Track Twenty-Eight", "Artist AB", "Electronic", 330000, 2600, 370, 92},
		{"29", "Track Twenty-Nine", "Artist AC", "Ambient", 440000, 3500, 540, 135},
		{"30", "Track Thirty", "Artist AD", "Techno", 370000, 3000, 440, 110},
		{"31", "Track Thirty-One", "Artist AE", "House", 340000, 2700, 390, 98},
		{"32", "Track Thirty-Two", "Artist AF", "Electronic", 390000, 3200, 490, 122},
		{"33", "Track Thirty-Three", "Artist AG", "Ambient", 410000, 3300, 510, 128},
		{"34", "Track Thirty-Four", "Artist AH", "Techno", 380000, 3050, 460, 115},
		{"35", "Track Thirty-Five", "Artist AI", "House", 350000, 2850, 410, 102},
	}

	return m.renderTracks(mockTracks, durationMin, searchQuery, offset, limit)
}

func (m *MockDataService) GetFavoritesTracks(durationMin int, searchQuery string, offset int, limit int) string {
	mockFavorites := []MockTrack{
		{"101", "Favorite Track One", "Artist X", "Electronic", 180000, 1500, 250, 50},
		{"102", "Another Favorite", "Artist Y", "House", 240000, 3200, 480, 120},
		{"103", "Liked Song", "Artist Z", "Techno", 60000, 890, 120, 25},
		{"104", "Track Four", "Artist A", "House", 200000, 2100, 300, 80},
		{"105", "Track Five", "Artist B", "Electronic", 150000, 1800, 220, 45},
		{"106", "Track Six", "Artist C", "Techno", 300000, 4500, 600, 150},
		{"107", "Track Seven", "Artist D", "Ambient", 400000, 3200, 400, 100},
		{"108", "Track Eight", "Artist E", "House", 180000, 2800, 350, 90},
		{"109", "Track Nine", "Artist F", "Electronic", 220000, 1900, 250, 60},
		{"110", "Track Ten", "Artist G", "Techno", 250000, 3100, 420, 110},
		{"111", "Track Eleven", "Artist H", "House", 170000, 2200, 280, 70},
		{"112", "Track Twelve", "Artist I", "Electronic", 190000, 2400, 320, 85},
		{"113", "Track Thirteen", "Artist J", "Ambient", 350000, 3800, 500, 130},
		{"114", "Track Fourteen", "Artist K", "Techno", 280000, 4100, 550, 140},
		{"115", "Track Fifteen", "Artist L", "House", 160000, 1600, 200, 40},
		{"116", "Track Sixteen", "Artist M", "Electronic", 210000, 2700, 360, 95},
		{"117", "Track Seventeen", "Artist N", "Ambient", 420000, 3400, 450, 120},
		{"118", "Track Eighteen", "Artist O", "Techno", 230000, 3000, 400, 100},
		{"119", "Track Nineteen", "Artist P", "House", 195000, 2300, 300, 75},
		{"120", "Track Twenty", "Artist Q", "Electronic", 205000, 2600, 340, 88},
		{"121", "Track Twenty-One", "Artist R", "Ambient", 380000, 3600, 480, 125},
		{"122", "Track Twenty-Two", "Artist S", "Techno", 270000, 4000, 530, 135},
		{"123", "Track Twenty-Three", "Artist T", "House", 185000, 2500, 330, 82},
		{"124", "Track Twenty-Four", "Artist U", "Electronic", 215000, 2850, 380, 98},
		{"125", "Track Twenty-Five", "Artist V", "Ambient", 360000, 3500, 460, 115},
		{"126", "Track Twenty-Six", "Artist W", "Techno", 260000, 3900, 520, 132},
		{"127", "Track Twenty-Seven", "Artist X", "House", 175000, 2400, 310, 78},
		{"128", "Track Twenty-Eight", "Artist Y", "Electronic", 200000, 2700, 360, 92},
		{"129", "Track Twenty-Nine", "Artist Z", "Ambient", 400000, 3700, 490, 128},
		{"130", "Track Thirty", "Artist A", "Techno", 245000, 3300, 440, 112},
		{"131", "Long Track One", "Artist B", "Ambient", 780000, 500, 100, 20},
		{"132", "Long Track Two", "Artist C", "Ambient", 900000, 600, 150, 30},
		{"133", "Long Track Three", "Artist D", "Ambient", 840000, 450, 80, 15},
		{"134", "Long Track Four", "Artist E", "Ambient", 1200000, 800, 200, 50},
		{"135", "Long Track Five", "Artist F", "Ambient", 960000, 700, 180, 40},
		{"136", "Long Track Six", "Artist G", "Ambient", 1080000, 900, 250, 60},
		{"137", "Long Track Seven", "Artist H", "Ambient", 720000, 400, 70, 10},
		{"138", "Long Track Eight", "Artist I", "Ambient", 1140000, 850, 220, 55},
		{"139", "Long Track Nine", "Artist J", "Ambient", 660000, 350, 60, 8},
		{"140", "Long Track Ten", "Artist K", "Ambient", 1320000, 1000, 300, 75},
		{"141", "Medium-Long One", "Artist L", "Electronic", 480000, 1500, 200, 45},
		{"142", "Medium-Long Two", "Artist M", "House", 540000, 1800, 280, 60},
		{"143", "Medium-Long Three", "Artist N", "Techno", 510000, 1600, 240, 52},
		{"144", "Medium-Long Four", "Artist O", "Ambient", 570000, 2000, 320, 70},
		{"145", "Medium-Long Five", "Artist P", "Electronic", 450000, 1400, 180, 38},
	}

	return m.renderTracks(mockFavorites, durationMin, searchQuery, offset, limit)
}

func (m *MockDataService) renderTracks(tracks []MockTrack, durationMin int, searchQuery string, offset int, limit int) string {
	var sb strings.Builder
	sb.WriteString("<div class=\"stream-flip-grid\">")

	searchLower := strings.ToLower(searchQuery)

	count := 0
	loaded := 0

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

		// Apply pagination
		if loaded < offset {
			loaded++
			continue
		}

		if count >= limit {
			break
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
		count++
		loaded++
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
