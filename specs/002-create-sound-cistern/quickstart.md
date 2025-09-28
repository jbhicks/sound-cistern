# Quickstart: Sound Cistern

## Setup
1. Clone and setup my-go-saas-template.
2. Configure .env with Soundcloud Client ID and Secret.
3. Run database migrations.
4. Start the Buffalo server.

## Test Scenarios
1. **Authentication**: Visit /auth/soundcloud, login with Soundcloud, verify redirect to feed.
2. **Feed Display**: After login, visit /feed, verify tracks load from database in <2s.
3. **Filtering**: Use filter bar to filter by length, genre, post time; verify real-time filtering.
4. **Error Handling**: Disconnect internet, visit /feed, verify cached data with warning.
5. **Performance**: Load feed, measure time <2s.

## Validation
- All acceptance scenarios pass.
- No common exploits in auth.
- Data retained for 2 weeks.