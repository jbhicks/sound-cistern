# Data Model: Sound Cistern

## Entities

### User
- id: UUID, primary key
- soundcloud_id: String, unique
- access_token: String, encrypted
- created_at: Timestamp
- Relationships: Has many Tracks, has one Feed

### Track
- id: UUID, primary key
- user_id: UUID, foreign key to User
- soundcloud_id: String, unique per user
- title: String
- length: Integer (seconds)
- genre: String (array or comma-separated)
- post_time: Timestamp
- created_at: Timestamp
- Relationships: Belongs to User

### Feed
- id: UUID, primary key
- user_id: UUID, foreign key to User
- tracks: JSON array of Track IDs or embedded
- updated_at: Timestamp
- Relationships: Belongs to User

## Validation Rules
- User soundcloud_id must be unique.
- Track length must be positive.
- Feed updated_at must be recent for cache validity.

## State Transitions
- User: Created on Soundcloud auth.
- Track: Added on feed fetch, updated on refresh.
- Feed: Refreshed periodically, trimmed to 2 weeks old tracks.