(function() {
    'use strict';

    const STORAGE_KEY = 'soundcistern_play_history';
    const MAX_PLAYS = 1000;
    const DATA_VERSION = 1;

    function getStorageData() {
        try {
            const data = localStorage.getItem(STORAGE_KEY);
            if (!data) {
                return { version: DATA_VERSION, plays: [] };
            }
            const parsed = JSON.parse(data);
            if (!parsed || !Array.isArray(parsed.plays)) {
                return { version: DATA_VERSION, plays: [] };
            }
            return parsed;
        } catch (e) {
            console.error('Error reading analytics data:', e);
            return { version: DATA_VERSION, plays: [] };
        }
    }

    function saveStorageData(data) {
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
            return true;
        } catch (e) {
            console.error('Error saving analytics data:', e);
            return false;
        }
    }

    function recordPlay(track) {
        if (!track || !track.id) {
            console.warn('Invalid track object provided to recordPlay');
            return false;
        }

        const data = getStorageData();

        const playRecord = {
            track_id: track.id,
            title: track.title || 'Unknown Track',
            artist: track.artist || 'Unknown Artist',
            genre: track.genre || 'Unknown',
            duration: track.duration || 0,
            played_at: new Date().toISOString()
        };

        data.plays.push(playRecord);

        while (data.plays.length > MAX_PLAYS) {
            data.plays.shift();
        }

        return saveStorageData(data);
    }

    function getTopTracks(limit = 10) {
        const data = getStorageData();
        const trackCounts = {};

        data.plays.forEach(play => {
            const key = play.track_id;
            if (!trackCounts[key]) {
                trackCounts[key] = {
                    track_id: play.track_id,
                    title: play.title,
                    artist: play.artist,
                    genre: play.genre,
                    duration: play.duration,
                    play_count: 0
                };
            }
            trackCounts[key].play_count++;
        });

        const sorted = Object.values(trackCounts)
            .sort((a, b) => b.play_count - a.play_count)
            .slice(0, limit);

        return sorted;
    }

    function getTopArtists(limit = 10) {
        const data = getStorageData();
        const artistCounts = {};

        data.plays.forEach(play => {
            const artist = play.artist;
            if (!artistCounts[artist]) {
                artistCounts[artist] = {
                    artist: artist,
                    play_count: 0,
                    tracks: new Set()
                };
            }
            artistCounts[artist].play_count++;
            artistCounts[artist].tracks.add(play.track_id);
        });

        const sorted = Object.values(artistCounts)
            .map(item => ({
                artist: item.artist,
                play_count: item.play_count,
                unique_tracks: item.tracks.size
            }))
            .sort((a, b) => b.play_count - a.play_count)
            .slice(0, limit);

        return sorted;
    }

    function getGenreDistribution() {
        const data = getStorageData();
        const genreCounts = {};
        let total = 0;

        data.plays.forEach(play => {
            const genre = play.genre || 'Unknown';
            if (!genreCounts[genre]) {
                genreCounts[genre] = 0;
            }
            genreCounts[genre]++;
            total++;
        });

        const distribution = Object.entries(genreCounts)
            .map(([genre, count]) => ({
                genre: genre,
                play_count: count,
                percentage: total > 0 ? Math.round((count / total) * 100 * 10) / 10 : 0
            }))
            .sort((a, b) => b.play_count - a.play_count);

        return distribution;
    }

    function getListeningTime() {
        const data = getStorageData();
        let totalMs = 0;

        data.plays.forEach(play => {
            totalMs += play.duration || 0;
        });

        const totalSeconds = totalMs / 1000;
        const totalMinutes = totalSeconds / 60;
        const totalHours = totalMinutes / 60;

        return {
            totalMilliseconds: totalMs,
            totalSeconds: Math.round(totalSeconds),
            totalMinutes: Math.round(totalMinutes * 10) / 10,
            totalHours: Math.round(totalHours * 100) / 100,
            totalPlays: data.plays.length
        };
    }

    function getPlayHistory(limit = 100) {
        const data = getStorageData();
        const plays = data.plays.slice(-limit).reverse();
        return plays;
    }

    function clearHistory() {
        try {
            localStorage.removeItem(STORAGE_KEY);
            return true;
        } catch (e) {
            console.error('Error clearing analytics data:', e);
            return false;
        }
    }

    window.SoundCisternAnalytics = {
        recordPlay: recordPlay,
        getTopTracks: getTopTracks,
        getTopArtists: getTopArtists,
        getGenreDistribution: getGenreDistribution,
        getListeningTime: getListeningTime,
        getPlayHistory: getPlayHistory,
        clearHistory: clearHistory
    };
})();
