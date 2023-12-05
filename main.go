package main

import (
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/interfaces"
	"github.com/SGE-AI/sge-bot/logger"
	"github.com/SGE-AI/sge-bot/slackapi"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

type Application struct {
	config config.Config
	logger logger.Logger
	socket interfaces.SocketConnection
	slack  slackapi.SlackAPI
}

func ProvideApplication(
	config config.Config,
	logger logger.Logger,
	socket interfaces.SocketConnection,
	slack slackapi.SlackAPI,
) *Application {
	return &Application{
		config: config,
		logger: logger,
		socket: socket,
		slack:  slack,
	}
}

func handleRequests(port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	_ = godotenv.Load()

	var app *Application
	app = initializeApp()

	botUser, err := app.slack.GetBotUserId()
	if err != nil {
		app.logger.Log(logger.ERROR, err.Error())
		os.Exit(1)
	}
	app.config.SetBotUserID(botUser)
	app.logger.Log(logger.INFO, "bot user id: %s", app.config.BotUserID())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app.logger.Log(logger.INFO, "http listening on port %s", port)
	go handleRequests(":" + port)

	err = app.socket.Run()
	if err != nil {
		app.logger.Log(logger.ERROR, err.Error())
	}
}
