package usecase

import (
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/SGE-AI/sge-bot/repository"
	"time"
)

type Statistics interface {
	UsedBy(SlackUserID string)
}

type statistics struct {
	spreadSheetRepo repository.SpreadsheetRepository
	log             logger.Logger
}

func (s *statistics) UsedBy(SlackUserID string) {
	if s.spreadSheetRepo == nil {
		return
	}

	stat, err := s.spreadSheetRepo.Get(SlackUserID)
	if err != nil {
		s.log.Log(logger.ERROR, "failed to get statistics: %v", err)
		return
	}

	s.log.Log(logger.INFO, "get statistics: user_id=%s, use_count=%d, last_used=%s", stat.SlackUserID, stat.UseCount, stat.LastUsed)

	stat.UseCount++
	stat.LastUsed = time.Now().Format(time.RFC3339)
	err = s.spreadSheetRepo.Update(stat)
	if err != nil {
		s.log.Log(logger.ERROR, "failed to update statistics: %v", err)
		return
	}
	s.log.Log(logger.INFO, "update statistics: user_id=%s, use_count=%d, last_used=%s", stat.SlackUserID, stat.UseCount, stat.LastUsed)
}

func ProvideStatistics(spreadSheetRepo repository.SpreadsheetRepository, log logger.Logger) Statistics {
	return &statistics{
		spreadSheetRepo: spreadSheetRepo,
		log:             log,
	}
}
