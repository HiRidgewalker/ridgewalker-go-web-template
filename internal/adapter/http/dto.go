package httpadapter

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/limber/learning-go-web-notebook/internal/domain"
)

// 请求 DTO -------------------------------------------------------------------

type createNotebookRequest struct {
	Title   string `json:"title" validate:"required,max=200"`
	Content string `json:"content" validate:"max=20000"`
}

type updateNotebookRequest struct {
	Title   *string `json:"title" validate:"omitempty,max=200"`
	Content *string `json:"content" validate:"omitempty,max=20000"`
}

// 响应 DTO -------------------------------------------------------------------

type notebookResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toNotebookResponse(notebook *domain.Notebook) notebookResponse {
	return notebookResponse{
		ID:        notebook.ID,
		Title:     notebook.Title,
		Content:   notebook.Content,
		CreatedAt: notebook.CreatedAt,
		UpdatedAt: notebook.UpdatedAt,
	}
}

type dataResponse struct {
	Data any `json:"data"`
}

type listResponse struct {
	Data []notebookResponse `json:"data"`
	Meta listMeta           `json:"meta"`
}

type listMeta struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}

// 辅助函数 -------------------------------------------------------------------

// NewValidator 构造传输层 validator。错误详情中的字段名使用 JSON 名称，而不是
// Go 结构体字段名。
func NewValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return v
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
