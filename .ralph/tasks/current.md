# Sound Cistern Development Tasks

## Current Task: soundcloud-mvp
**Status**: In Progress  
**Started**: 2025-01-25  
**Iteration**: 1

### Success Criteria
- [ ] OAuth Authentication (OAuth 2.1 with PKCE)
- [ ] Stream Display (chronological track listing)
- [ ] Track Metadata (title, artist, duration, artwork)
- [ ] Basic Filtering (date, genre filters)
- [ ] Persistent State (remember last played)

### Implementation Progress
- [x] PocketBase migrations for soundcloud_users
- [x] Basic OAuth URL generation
- [ ] OAuth callback handling
- [ ] Track fetching from Soundcloud API
- [ ] Stream display template
- [ ] Filtering UI components

### Blockers
- [ ] OAuth token refresh mechanism
- [ ] API rate limiting handling
- [ ] Error handling for API failures

---

## Backlog Tasks (Future Iterations)

### soundcloud-mvp-plus
- Favorites system
- Search integration
- Playlist creation
- RSS feed generation

### soundcloud-social
- Follow/unfollow users
- Like/unlike tracks
- Comment system
- Activity feed

### soundcloud-analytics
- Listening statistics
- Track analytics
- User engagement metrics
- Export data (CSV, JSON)

### soundcloud-mobile
- PWA support
- Offline playback
- Mobile-responsive design
- App store deployment