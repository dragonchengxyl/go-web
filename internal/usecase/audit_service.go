package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/audit"
	"github.com/studio/platform/internal/pkg/apperr"
)

// AuditService handles audit logging
type AuditService struct {
	auditRepo audit.Repository
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo audit.Repository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// LogInput represents input for creating an audit log
type LogInput struct {
	UserID       *uuid.UUID
	Username     string
	Action       audit.Action
	Resource     audit.Resource
	ResourceID   *uuid.UUID
	IPAddress    string
	UserAgent    string
	BeforeData   interface{}
	AfterData    interface{}
	ErrorMessage *string
}

// Log creates a new audit log entry
func (s *AuditService) Log(ctx context.Context, input LogInput) error {
	var beforeJSON, afterJSON *string

	if input.BeforeData != nil {
		data, err := json.Marshal(input.BeforeData)
		if err == nil {
			str := string(data)
			beforeJSON = &str
		}
	}

	if input.AfterData != nil {
		data, err := json.Marshal(input.AfterData)
		if err == nil {
			str := string(data)
			afterJSON = &str
		}
	}

	log := &audit.Log{
		ID:           uuid.New(),
		UserID:       input.UserID,
		Username:     input.Username,
		Action:       input.Action,
		Resource:     input.Resource,
		ResourceID:   input.ResourceID,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		BeforeData:   beforeJSON,
		AfterData:    afterJSON,
		ErrorMessage: input.ErrorMessage,
		CreatedAt:    time.Now(),
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		// Don't fail the main operation if audit logging fails
		// Just log the error (in production, send to monitoring system)
		return nil
	}

	return nil
}

// ListAuditLogsInput represents input for listing audit logs
type ListAuditLogsInput struct {
	UserID     *uuid.UUID
	Action     *audit.Action
	Resource   *audit.Resource
	ResourceID *uuid.UUID
	StartTime  *string
	EndTime    *string
	Page       int
	PageSize   int
}

// ListAuditLogsOutput represents output for listing audit logs
type ListAuditLogsOutput struct {
	Logs  []*audit.Log `json:"logs"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

// ListAuditLogs retrieves audit logs with filters
func (s *AuditService) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (*ListAuditLogsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	filter := audit.ListFilter{
		UserID:     input.UserID,
		Action:     input.Action,
		Resource:   input.Resource,
		ResourceID: input.ResourceID,
		StartTime:  input.StartTime,
		EndTime:    input.EndTime,
		Page:       input.Page,
		PageSize:   input.PageSize,
	}

	logs, total, err := s.auditRepo.List(ctx, filter)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询审计日志失败", err)
	}

	return &ListAuditLogsOutput{
		Logs:  logs,
		Total: total,
		Page:  input.Page,
		Size:  len(logs),
	}, nil
}

// GetUserAuditLogs retrieves audit logs for a specific user
func (s *AuditService) GetUserAuditLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*audit.Log, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	logs, err := s.auditRepo.GetByUserID(ctx, userID, limit)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询用户审计日志失败", err)
	}

	return logs, nil
}

// GetResourceAuditLogs retrieves audit logs for a specific resource
func (s *AuditService) GetResourceAuditLogs(ctx context.Context, resource audit.Resource, resourceID uuid.UUID, limit int) ([]*audit.Log, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	logs, err := s.auditRepo.GetByResource(ctx, resource, resourceID, limit)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "查询资源审计日志失败", err)
	}

	return logs, nil
}
