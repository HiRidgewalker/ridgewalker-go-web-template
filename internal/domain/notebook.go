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

// Notebook 是 notebook 限界上下文的聚合根。
//
// 它刻意只使用标准库类型，使 domain 不依赖 HTTP、SQL 与日志；ID 与时间戳在
// 实体首次写入时由持久化 adapter 赋值。
type Notebook struct {
	ID        string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewNotebook 校验并构造 notebook。
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

// Update 执行部分更新：nil 字段保持不变。
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
