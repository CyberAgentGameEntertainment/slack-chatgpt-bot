//go:build wireinject
// +build wireinject

package main

import (
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/gpt"
	"github.com/SGE-AI/sge-bot/interfaces"
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/SGE-AI/sge-bot/repository"
	"github.com/SGE-AI/sge-bot/slackapi"
	"github.com/SGE-AI/sge-bot/usecase"
	"github.com/google/wire"
)

func initializeApp() *Application {
	wire.Build(
		config.ProvideConfig,
		logger.ProvideLogger,
		repository.ProvideContextCancelRepository,
		repository.ProvideSpreadsheetRepository,
		gpt.ProvideGPTClient,
		usecase.ProvideChat,
		usecase.ProvideStatistics,
		interfaces.ProvideEventHandler,
		interfaces.ProvideSocketConnection,
		slackapi.ProvideSlackAPI,
		ProvideApplication,
	)

	return nil
}
