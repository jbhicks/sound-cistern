package models

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Post represents a blog post
type Post struct {
	ID              int       `json:"id" db:"id" form:"-"`
	Title           string    `json:"title" db:"title"`
	Slug            string    `json:"slug" db:"slug"`
	Content         string    `json:"content" db:"content"`
	Excerpt         string    `json:"excerpt" db:"excerpt"`
	Published       bool      `json:"published" db:"published"`
	MetaTitle       string    `json:"meta_title" db:"meta_title"`
	MetaDescription string    `json:"meta_description" db:"meta_description"`
	MetaKeywords    string    `json:"meta_keywords" db:"meta_keywords"`
	OgTitle         string    `json:"og_title" db:"og_title"`
	OgDescription   string    `json:"og_description" db:"og_description"`
	OgImage         string    `json:"og_image" db:"og_image"`
	Image           string    `json:"image" db:"image"`
	ImageAlt        string    `json:"image_alt" db:"image_alt"`
	AuthorID        uuid.UUID `json:"author_id" db:"author_id" form:"-"`
	Author          *User     `json:"author,omitempty" belongs_to:"user" fk_id:"author_id" form:"-"`
	CreatedAt       time.Time `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at" form:"-"`
}

// String is not required by pop and may be deleted
func (p Post) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Posts is not required by pop and may be deleted
type Posts []Post

// String is not required by pop and may be deleted
func (p Posts) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
func (p *Post) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: p.Title, Name: "Title"},
		&validators.StringIsPresent{Field: p.Content, Name: "Content"},
		&validators.StringIsPresent{Field: p.Slug, Name: "Slug"},
		&validators.UUIDIsPresent{Field: p.AuthorID, Name: "AuthorID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (p *Post) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (p *Post) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// GenerateSlug creates a URL-friendly slug from the title
func (p *Post) GenerateSlug() {
	if p.Slug == "" && p.Title != "" {
		// Convert to lowercase and replace spaces/special chars with hyphens
		slug := strings.ToLower(p.Title)
		// Replace non-alphanumeric characters with hyphens
		reg := regexp.MustCompile(`[^a-z0-9]+`)
		slug = reg.ReplaceAllString(slug, "-")
		// Remove leading/trailing hyphens
		slug = strings.Trim(slug, "-")
		p.Slug = slug
	}
}

// BeforeCreate runs before creating a post
func (p *Post) BeforeCreate(tx *pop.Connection) error {
	p.GenerateSlug()
	return nil
}

// BeforeUpdate runs before updating a post
func (p *Post) BeforeUpdate(tx *pop.Connection) error {
	// Only regenerate slug if it's empty
	if p.Slug == "" {
		p.GenerateSlug()
	}
	return nil
}

// GenerateExcerpt creates an excerpt from content if not provided
func (p *Post) GenerateExcerpt() {
	if p.Excerpt == "" && p.Content != "" {
		content := strings.TrimSpace(p.Content)
		if len(content) > 200 {
			p.Excerpt = content[:200] + "..."
		} else {
			p.Excerpt = content
		}
	}
}
