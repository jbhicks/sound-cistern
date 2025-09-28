# Strapi CMS Integration

This project now uses Strapi as a headless CMS for content management, replacing the previous custom admin interface.

## Quick Start

1. **Start Strapi and PostgreSQL**:
   ```bash
   docker-compose -f docker-compose.strapi.yml up -d
   ```

2. **Set up Strapi Admin**:
   - Open http://localhost:1337/admin
   - Create your first admin user
   - The setup wizard will guide you through initial configuration

3. **Configure Post Content Type**:
   - In Strapi admin, go to Content-Types Builder
   - Create a new Collection Type called "Post"
   - Add these fields:
     - `title` (Text)
     - `slug` (Text, unique)
     - `content` (Rich Text)
     - `excerpt` (Text)
     - `meta_title` (Text)
     - `meta_description` (Text)
     - `published_at` (Date)
     - `featured_image` (Media)

4. **Configure API Permissions**:
   - Go to Settings > Roles > Public
   - Enable `find` and `findOne` for Post content type
   - This allows Buffalo to fetch published posts

## Development Workflow

### Content Management
- **Admin users**: Manage content at http://localhost:1337/admin
- **Content creation**: Use Strapi's intuitive interface instead of custom forms
- **Media management**: Upload and organize images through Strapi's media library
- **Publishing**: Use Strapi's draft/publish workflow

### API Integration
- **Buffalo app**: Fetches content from Strapi API (`http://localhost:1337/api/posts`)
- **Public routes**: `/blog` and `/blog/:slug` continue to work as before
- **Admin routes**: Redirect to Strapi admin interface

## Benefits

1. **Simplified Admin**: No more complex custom forms and validation
2. **Media Management**: Built-in image upload and organization
3. **User-Friendly**: Intuitive WYSIWYG editor and content management
4. **Professional Features**: Draft/publish, SEO fields, content versioning
5. **Maintainability**: Less custom code to maintain
6. **Scalability**: Strapi handles content scaling and performance

## API Endpoints

- `GET /api/posts` - List all published posts
- `GET /api/posts/:id` - Get specific post
- `GET /api/posts?filters[slug][$eq]=:slug` - Get post by slug

## Environment Variables

Add to your `.env` file:
```
STRAPI_URL=http://localhost:1337
STRAPI_API_TOKEN=your-api-token-here
```

## Production Deployment

For production:
1. Update JWT secrets and API tokens in docker-compose.strapi.yml
2. Configure proper database credentials
3. Set up SSL/TLS for Strapi admin interface
4. Configure proper CORS settings in Strapi
