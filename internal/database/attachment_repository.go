package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"todo-calendar/internal/models"
)

// AttachmentRepository 附件仓库
type AttachmentRepository struct {
	db *sql.DB
}

// NewAttachmentRepository 创建附件仓库实例
func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create 创建附件记录
func (r *AttachmentRepository) Create(attachment *models.Attachment) (int64, error) {
	query := `
		INSERT INTO attachments (todo_id, file_name, storage_path, file_size, mime_type, is_encrypted, encryption_key, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		attachment.TodoID,
		attachment.FileName,
		attachment.StoragePath,
		attachment.FileSize,
		attachment.MimeType,
		attachment.IsEncrypted,
		attachment.EncryptionKey,
		time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Delete 删除附件
func (r *AttachmentRepository) Delete(id int64) error {
	// 先获取附件信息
	attachment, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// 删除文件
	if err := os.Remove(attachment.StoragePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 删除记录
	_, err = r.db.Exec("DELETE FROM attachments WHERE id = ?", id)
	return err
}

// GetByID 根据ID获取附件
func (r *AttachmentRepository) GetByID(id int64) (*models.Attachment, error) {
	query := `
		SELECT id, todo_id, file_name, storage_path, file_size, mime_type, is_encrypted, encryption_key, created_at
		FROM attachments WHERE id = ?
	`
	attachment := &models.Attachment{}
	err := r.db.QueryRow(query, id).Scan(
		&attachment.ID,
		&attachment.TodoID,
		&attachment.FileName,
		&attachment.StoragePath,
		&attachment.FileSize,
		&attachment.MimeType,
		&attachment.IsEncrypted,
		&attachment.EncryptionKey,
		&attachment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return attachment, nil
}

// GetByTodoID 根据待办ID获取附件列表
func (r *AttachmentRepository) GetByTodoID(todoID int64) ([]models.Attachment, error) {
	query := `
		SELECT id, todo_id, file_name, storage_path, file_size, mime_type, is_encrypted, encryption_key, created_at
		FROM attachments WHERE todo_id = ?
	`
	rows, err := r.db.Query(query, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := []models.Attachment{}
	for rows.Next() {
		var attachment models.Attachment
		err := rows.Scan(
			&attachment.ID,
			&attachment.TodoID,
			&attachment.FileName,
			&attachment.StoragePath,
			&attachment.FileSize,
			&attachment.MimeType,
			&attachment.IsEncrypted,
			&attachment.EncryptionKey,
			&attachment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}
	return attachments, nil
}

// DeleteByTodoID 删除待办的所有附件
func (r *AttachmentRepository) DeleteByTodoID(todoID int64) error {
	// 先获取所有附件
	attachments, err := r.GetByTodoID(todoID)
	if err != nil {
		return err
	}

	// 删除文件
	for _, attachment := range attachments {
		os.Remove(attachment.StoragePath)
	}

	// 删除记录
	_, err = r.db.Exec("DELETE FROM attachments WHERE todo_id = ?", todoID)
	return err
}

// EncryptAndSaveFile 加密并保存文件
func (r *AttachmentRepository) EncryptAndSaveFile(todoID int64, fileName string, data []byte, mimeType string) (*models.Attachment, error) {
	// 生成32字节的AES密钥
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// 加密数据
	encryptedData, err := encryptAES(data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// 获取存储目录
	storageDir, err := getAttachmentDir()
	if err != nil {
		return nil, err
	}

	// 生成唯一文件名
	uniqueName := fmt.Sprintf("%d_%d_%s", todoID, time.Now().UnixNano(), fileName)
	storagePath := filepath.Join(storageDir, uniqueName+".enc")

	// 保存加密文件
	if err := os.WriteFile(storagePath, encryptedData, 0600); err != nil {
		return nil, fmt.Errorf("failed to save encrypted file: %w", err)
	}

	// 创建附件记录
	attachment := &models.Attachment{
		TodoID:        todoID,
		FileName:      fileName,
		StoragePath:   storagePath,
		FileSize:      int64(len(data)),
		MimeType:      mimeType,
		IsEncrypted:   true,
		EncryptionKey: base64.StdEncoding.EncodeToString(key),
	}

	id, err := r.Create(attachment)
	if err != nil {
		os.Remove(storagePath)
		return nil, err
	}
	attachment.ID = id

	return attachment, nil
}

// DecryptFile 解密文件
func (r *AttachmentRepository) DecryptFile(id int64) ([]byte, error) {
	attachment, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 读取加密文件
	encryptedData, err := os.ReadFile(attachment.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted file: %w", err)
	}

	if !attachment.IsEncrypted {
		return encryptedData, nil
	}

	// 解码密钥
	key, err := base64.StdEncoding.DecodeString(attachment.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	// 解密
	return decryptAES(encryptedData, key)
}

// getAttachmentDir 获取附件存储目录
func getAttachmentDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(filepath.Dir(exe), "data", "attachments")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// encryptAES AES加密
func encryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decryptAES AES解密
func decryptAES(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
