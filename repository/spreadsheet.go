package repository

import (
	"context"
	"fmt"
	"github.com/SGE-AI/sge-bot/config"
	"github.com/SGE-AI/sge-bot/logger"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"strconv"
)

type (
	UserStatistics struct {
		SlackUserID string
		UseCount    int
		LastUsed    string
	}

	SpreadsheetRepository interface {
		Get(SlackUserID string) (UserStatistics, error)
		Update(UserStatistics) error
	}

	spreadsheetRepository struct {
		service       *sheets.Service
		spreadSheetID string
	}
)

func (s *spreadsheetRepository) Get(SlackUserID string) (UserStatistics, error) {
	readRange := fmt.Sprintf("A:C")
	resp, err := s.service.Spreadsheets.Values.Get(s.spreadSheetID, readRange).Do()

	if err != nil {
		return UserStatistics{}, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	for _, row := range resp.Values {
		if len(row) >= 3 {
			userId, ok := row[0].(string)
			if !ok {
				continue
			}
			if userId == SlackUserID {
				useCount, _ := strconv.Atoi(row[1].(string))
				lastUsed, _ := row[2].(string)

				return UserStatistics{
					SlackUserID: userId,
					UseCount:    useCount,
					LastUsed:    lastUsed,
				}, nil
			}
		}
	}

	return UserStatistics{
		SlackUserID: SlackUserID,
		UseCount:    0,
		LastUsed:    "",
	}, nil
}

func (s *spreadsheetRepository) Update(stat UserStatistics) error {
	readRange := fmt.Sprintf("A:C")
	resp, err := s.service.Spreadsheets.Values.Get(s.spreadSheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	var rowToUpdate int = -1
	for idx, row := range resp.Values {
		if len(row) > 0 {
			userId, ok := row[0].(string)
			if ok && userId == stat.SlackUserID {
				rowToUpdate = idx
				break
			}
		}
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{{stat.SlackUserID, stat.UseCount, stat.LastUsed}},
	}

	if rowToUpdate >= 0 {
		updateRange := fmt.Sprintf("A%d:C%d", rowToUpdate+1, rowToUpdate+1)
		_, err = s.service.Spreadsheets.Values.Update(s.spreadSheetID, updateRange, valueRange).ValueInputOption("RAW").Do()
	} else {
		_, err = s.service.Spreadsheets.Values.Append(s.spreadSheetID, readRange, valueRange).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()
	}

	if err != nil {
		return fmt.Errorf("unable to update data from sheet: %v", err)
	}

	return nil
}

func NewSpreadsheetRepository(cfg config.Config) (SpreadsheetRepository, error) {
	jsonKey := cfg.GoogleApplicationCredentialsJSON()
	conf, err := google.JWTConfigFromJSON([]byte(jsonKey), sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	conf.Subject = cfg.GoogleServiceAccountEmail()
	client := conf.Client(context.Background())
	service, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets Client %v", err)
	}

	return &spreadsheetRepository{
		service:       service,
		spreadSheetID: cfg.SpreadSheetID(),
	}, nil
}

var spreadsheetRepoSingleton SpreadsheetRepository

func ProvideSpreadsheetRepository(cfg config.Config, log logger.Logger) SpreadsheetRepository {
	if spreadsheetRepoSingleton == nil {
		repo, err := NewSpreadsheetRepository(cfg)
		if err != nil {
			log.Log(logger.ERROR, "failed to create spreadsheet repository: %v", err)
			return nil
		}
		spreadsheetRepoSingleton = repo
	}
	return spreadsheetRepoSingleton
}
