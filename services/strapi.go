package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// StrapiPost represents a post from Strapi API
type StrapiPost struct {
	ID         int                  `json:"id"`
	Attributes StrapiPostAttributes `json:"attributes"`
}

type StrapiPostAttributes struct {
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Content         string     `json:"content"`
	Excerpt         string     `json:"excerpt"`
	MetaTitle       string     `json:"meta_title"`
	MetaDescription string     `json:"meta_description"`
	PublishedAt     *time.Time `json:"publishedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// StrapiResponse represents the API response structure
type StrapiResponse struct {
	Data []StrapiPost `json:"data"`
}

type StrapiSingleResponse struct {
	Data StrapiPost `json:"data"`
}

// StrapiService handles communication with Strapi CMS
type StrapiService struct {
	baseURL  string
	apiToken string
	client   *http.Client
}

// NewStrapiService creates a new Strapi service instance
func NewStrapiService() *StrapiService {
	baseURL := os.Getenv("STRAPI_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1337"
	}

	return &StrapiService{
		baseURL:  baseURL,
		apiToken: os.Getenv("STRAPI_API_TOKEN"),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// makeRequest performs HTTP request to Strapi API
func (s *StrapiService) makeRequest(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/%s", s.baseURL, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add authorization header if API token is provided
	if s.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strapi API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetPublishedPosts fetches all published posts from Strapi
func (s *StrapiService) GetPublishedPosts() ([]StrapiPost, error) {
	// Only get published posts, sorted by publication date (newest first)
	endpoint := "posts?filters[publishedAt][$notNull]=true&sort=publishedAt:desc"

	body, err := s.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response StrapiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// GetPostBySlug fetches a specific post by its slug
func (s *StrapiService) GetPostBySlug(slug string) (*StrapiPost, error) {
	// Encode the slug for URL safety
	encodedSlug := url.QueryEscape(slug)
	endpoint := fmt.Sprintf("posts?filters[slug][$eq]=%s&filters[publishedAt][$notNull]=true", encodedSlug)

	body, err := s.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response StrapiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("post with slug '%s' not found", slug)
	}

	return &response.Data[0], nil
}

// GetPostByID fetches a specific post by its ID
func (s *StrapiService) GetPostByID(id int) (*StrapiPost, error) {
	endpoint := fmt.Sprintf("posts/%d", id)

	body, err := s.makeRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var response StrapiSingleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// HealthCheck verifies that Strapi is accessible
func (s *StrapiService) HealthCheck() error {
	url := fmt.Sprintf("%s/api/posts?pagination[limit]=1", s.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach Strapi at %s: %v", s.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Strapi returned status %d", resp.StatusCode)
	}

	return nil
}
