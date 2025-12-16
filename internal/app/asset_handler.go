package app

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"todo-calendar/internal/database"
	"todo-calendar/internal/models"
)

// AttachmentHandler 处理附件请求的自定义处理器
type AttachmentHandler struct {
	attachmentRepo *database.AttachmentRepository
}

// NewAttachmentHandler 创建附件处理器
func NewAttachmentHandler(attachmentRepo *database.AttachmentRepository) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentRepo: attachmentRepo,
	}
}

// ServeHTTP 处理 /attachment/{todoId}/{fileName} 请求
func (h *AttachmentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 解析路径: /attachment/{todoId}/{fileName}
	path := strings.TrimPrefix(r.URL.Path, "/attachment/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid attachment path", http.StatusBadRequest)
		return
	}

	todoID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	fileName := parts[1]

	// 查找附件
	attachments, err := h.attachmentRepo.GetByTodoID(todoID)
	if err != nil {
		http.Error(w, "Failed to get attachments", http.StatusInternalServerError)
		return
	}

	var targetAttachment *models.Attachment
	for i := range attachments {
		if attachments[i].FileName == fileName {
			targetAttachment = &attachments[i]
			break
		}
	}

	if targetAttachment == nil {
		http.Error(w, "Attachment not found", http.StatusNotFound)
		return
	}

	// 解密附件数据
	data, err := h.attachmentRepo.DecryptFile(targetAttachment.ID)
	if err != nil {
		http.Error(w, "Failed to decrypt attachment", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", targetAttachment.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Header().Set("Cache-Control", "max-age=3600")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetAttachmentAsDataURL 获取附件的 data URL（供前端调用）
func (a *App) GetAttachmentAsDataURL(todoID int64, fileName string) (string, error) {
	attachments, err := a.attachmentRepo.GetByTodoID(todoID)
	if err != nil {
		return "", err
	}

	for _, att := range attachments {
		if att.FileName == fileName {
			data, err := a.attachmentRepo.DecryptFile(att.ID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("data:%s;base64,%s", att.MimeType, base64.StdEncoding.EncodeToString(data)), nil
		}
	}

	return "", fmt.Errorf("attachment not found: %s", fileName)
}
