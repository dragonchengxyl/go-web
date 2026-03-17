package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/sponsor"
	"github.com/studio/platform/internal/pkg/apperr"
)

// SponsorSettingsService resolves runtime sponsor settings with DB override + config fallback.
type SponsorSettingsService struct {
	repo sponsor.Repository
	cfg  *configs.Config
}

func NewSponsorSettingsService(repo sponsor.Repository, cfg *configs.Config) *SponsorSettingsService {
	return &SponsorSettingsService{repo: repo, cfg: cfg}
}

func (s *SponsorSettingsService) Get(ctx context.Context) (configs.SponsorConfig, error) {
	if s.repo == nil {
		return s.cfg.Sponsor, nil
	}
	item, err := s.repo.Get(ctx)
	if err != nil {
		return configs.SponsorConfig{}, apperr.Wrap(apperr.CodeInternalError, "读取赞助配置失败", err)
	}
	if item == nil {
		return s.cfg.Sponsor, nil
	}
	return *item, nil
}

func (s *SponsorSettingsService) Update(ctx context.Context, cfg configs.SponsorConfig, updatedBy *uuid.UUID) (configs.SponsorConfig, error) {
	if cfg.MonthlyGoal < 0 || cfg.CurrentRaised < 0 {
		return configs.SponsorConfig{}, apperr.BadRequest("金额不能为负数")
	}

	if s.repo != nil {
		if err := s.repo.Upsert(ctx, cfg, updatedBy); err != nil {
			return configs.SponsorConfig{}, apperr.Wrap(apperr.CodeInternalError, "保存赞助配置失败", err)
		}
	}

	s.cfg.Sponsor = cfg
	return cfg, nil
}
