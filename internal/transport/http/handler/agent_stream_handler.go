package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
)

func (h *AgentHandler) StreamRun(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "当前环境不支持流式响应", nil))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher.Flush()

	writeEvent := func(event string, payload any) error {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	events, unsubscribe := h.service.SubscribeRun(runID)
	defer unsubscribe()

	detail, err := h.service.GetRunDetail(ctx, userID, runID)
	if err != nil {
		_ = writeEvent("error", gin.H{"message": err.Error()})
		return
	}
	if err := writeEvent("snapshot", detail); err != nil {
		return
	}

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := writeEvent("ping", gin.H{"ok": true}); err != nil {
				return
			}
		case event, ok := <-events:
			if !ok {
				return
			}
			nextDetail, err := h.service.GetRunDetail(ctx, userID, runID)
			if err != nil {
				_ = writeEvent("error", gin.H{"message": err.Error()})
				return
			}
			if err := writeEvent("update", event); err != nil {
				return
			}
			if err := writeEvent("snapshot", nextDetail); err != nil {
				return
			}
			if nextDetail.Run != nil {
				switch nextDetail.Run.Status {
				case "completed", "failed", "cancelled":
					_ = writeEvent("done", gin.H{"status": nextDetail.Run.Status})
					return
				}
			}
		}
	}
}
