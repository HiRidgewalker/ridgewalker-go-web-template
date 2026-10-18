package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxTitleLength   = 200
	MaxContentLength = 20000
)

// Notebook is the aggregate root of the notebook bounded context.
//
// It deliberately contains only standard-library types so the domain stays
// independent from HTTP, SQL and logging concerns. ID and timestamps are
// assigned by the persistence adapter when the entity is first stored.
type Notebook struct {
	ID        string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewNotebook validates and constructs a notebook.
func NewNotebook(title, content string) (*Notebook, error) {
	notebook := &Notebook{
		Title:   strings.TrimSpace(title),
		Content: content,
	}
	if err := notebook.validate(); err != nil {
		return nil, err
	}
	return notebook, nil
}

// Update applies a partial update: nil fields are left untouched.
func (n *Notebook) Update(title, content *string) error {
	if title != nil {
		n.Title = strings.TrimSpace(*title)
	}
	if content != nil {
		n.Content = *content
	}
	if err := n.validate(); err != nil {
		return err
	}
	n.UpdatedAt = time.Now().UTC()
	return nil
}

func (n *Notebook) validate() error {
	if n.Title == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if utf8.RuneCountInString(n.Title) > MaxTitleLength {
		return fmt.Errorf("%w: title must be at most %d characters", ErrValidation, MaxTitleLength)
	}
	if utf8.RuneCountInString(n.Content) > MaxContentLength {
		return fmt.Errorf("%w: content must be at most %d characters", ErrValidation, MaxContentLength)
	}
	return nil
}
