package repo

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/abobacode/tvparser/models"
)

type Repo struct{}

func (r *Repo) ConvertCustomDate(layout, input string) (time.Time, error) {
	newDate, err := time.Parse(layout, input)
	if err != nil {
		return time.Time{}, err
	}

	return newDate, nil
}

func (r *Repo) MakeRequest(URL string) (*http.Response, error) {
	response, err := http.Get(URL)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *Repo) GetIsoWithUTC(allPrograms []models.Programs, tz *time.Location) error {
	for i := range allPrograms {
		for j := range allPrograms[i].Program {
			dateDay := allPrograms[i].DayInt
			timeStartStr := allPrograms[i].Program[j].Time.Start
			timeFinishStr := allPrograms[i].Program[j].Time.Finish

			timeStart, err := time.Parse("15:04", timeStartStr)
			if err != nil {
				return err
			}

			timeFinish, err := time.Parse("15:04", timeFinishStr)
			if err != nil {
				return err
			}

			isoForStart := time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeStart.Hour(), timeStart.Minute(), 0, 0, tz)
			isoForFinish := time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeFinish.Hour(), timeFinish.Minute(), 0, 0, tz)

			allPrograms[i].Program[j].Time.StartISO = isoForStart.Format(time.RFC3339)
			allPrograms[i].Program[j].Time.FinishISO = isoForFinish.Format(time.RFC3339)
		}
	}

	return nil
}

func (r *Repo) GetTimeTransition(tvPrograms []models.Programs, timeZone *time.Location) error {
	for i := 0; i < len(tvPrograms)-1; i++ {
		for j := range tvPrograms[i].Program {
			dateDay := tvPrograms[i].DayInt
			dateDayNext := tvPrograms[i+1].DayInt
			timeStartStr := tvPrograms[i].Program[j].Time.Start
			timeFinishStr := tvPrograms[i].Program[j].Time.Finish

			timeStart, err := time.Parse("15:04", timeStartStr)
			if err != nil {
				log.Fatal(err)
			}

			timeFinish, err := time.Parse("15:04", timeFinishStr)
			if err != nil {
				log.Fatal(err)
			}

			var (
				isoForStart  time.Time
				isoForFinish time.Time
			)

			if j == len(tvPrograms[i].Program)-1 {
				isoForStart = time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeStart.Hour(), timeStart.Minute(), 0, 0, timeZone)
				isoForFinish = time.Date(dateDayNext.Year(), dateDayNext.Month(), dateDayNext.Day(), timeFinish.Hour(), timeFinish.Minute(), 0, 0, timeZone)
			} else {
				isoForStart = time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeStart.Hour(), timeStart.Minute(), 0, 0, timeZone)
				isoForFinish = time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeFinish.Hour(), timeFinish.Minute(), 0, 0, timeZone)
			}

			tvPrograms[i].Program[j].Time.StartISO = isoForStart.Format(time.RFC3339)
			tvPrograms[i].Program[j].Time.FinishISO = isoForFinish.Format(time.RFC3339)
		}
	}

	for i := len(tvPrograms) - 1; i < len(tvPrograms); i++ {
		for j := range tvPrograms[i].Program {
			dateDay := tvPrograms[i].DayInt
			tomorrow := dateDay.AddDate(0, 0, 1)
			timeStartStr := tvPrograms[i].Program[j].Time.Start
			timeFinishStr := tvPrograms[i].Program[j].Time.Finish

			timeStart, err := time.Parse("15:04", timeStartStr)
			if err != nil {
				log.Fatal(err)
			}

			timeFinish, err := time.Parse("15:04", timeFinishStr)
			if err != nil {
				log.Fatal(err)
			}

			var (
				isoForStart  time.Time
				isoForFinish time.Time
			)

			if j == len(tvPrograms[i].Program)-1 {
				isoForStart = time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeStart.Hour(), timeStart.Minute(), 0, 0, timeZone)
				isoForFinish = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), timeFinish.Hour(), timeFinish.Minute(), 0, 0, timeZone)
			} else {
				isoForStart = time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeStart.Hour(), timeStart.Minute(), 0, 0, timeZone)
				isoForFinish = time.Date(dateDay.Year(), dateDay.Month(), dateDay.Day(), timeFinish.Hour(), timeFinish.Minute(), 0, 0, timeZone)
			}

			tvPrograms[i].Program[j].Time.StartISO = isoForStart.Format(time.RFC3339)
			tvPrograms[i].Program[j].Time.FinishISO = isoForFinish.Format(time.RFC3339)
		}
	}

	return nil
}

func (r *Repo) GetCSV(pathToFile string, allPrograms []models.Programs) error {
	filePath := pathToFile

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'

	defer writer.Flush()

	headers := []string{
		"datetime_start",
		"datetime_finish",
		"title",
		"description",
		"channel",
		"channel_logo_url",
		"available_archive",
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, program := range allPrograms {
		for _, timeTitle := range program.Program {
			record := []string{
				timeTitle.Time.StartISO,
				timeTitle.Time.FinishISO,
				timeTitle.Title,
				timeTitle.Description,
				timeTitle.Channel,
				timeTitle.ChannelLogoURL,
				strconv.Itoa(timeTitle.AvailableArchive),
			}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

func New() *Repo {
	return &Repo{}
}
