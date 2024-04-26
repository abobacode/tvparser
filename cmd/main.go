package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/urfave/cli/v2"

	"github.com/abobacode/tvparser/models"
	"github.com/abobacode/tvparser/repo"
)

const (
	URL              = "https://tvschedule.today/in/tv-schedule/hare-krsna"
	channelName      = "Hare Krsna"
	channelLogoURL   = ""
	availableArchive = 0
)

type Fetcher interface {
	GetCSV(pathToFile string, allPrograms []models.Programs) error
	MakeRequest(URL string) (*http.Response, error)
	ConvertCustomDate(layout, input string) (time.Time, error)
	GetIsoWithUTC(allPrograms []models.Programs, tz *time.Location) error
	GetTimeTransition(tvPrograms []models.Programs, timeZone *time.Location) error
}

func main() {
	application := cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "output",
				Required: false,
				Value:    "epg.csv",
				Usage:    "Hare Krsna",
			},
		},
		Action: Main,
	}

	if err := application.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Main(ctx *cli.Context) error {
	var (
		allPrograms []models.Programs
		fetcher     Fetcher = repo.New()
		programs    []models.TimeTitle
		timeZone    = time.FixedZone("UTC+5:30", 5*60*60+30*60)
	)

	tvPrograms, err := GetPrograms(programs, allPrograms, fetcher)
	if err != nil {
		return err
	}

	if err := fetcher.GetTimeTransition(tvPrograms, timeZone); err != nil {
		return err
	}

	if err := fetcher.GetCSV(ctx.String("output"), tvPrograms); err != nil {
		return err
	}

	return nil
}

func GetPrograms(programs []models.TimeTitle, all []models.Programs, f Fetcher) ([]models.Programs, error) {
	response, err := f.MakeRequest(URL)
	if err != nil {
		log.Fatal(err)
	}

	days, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	days.Find("option").Each(func(i int, d *goquery.Selection) {
		value, _ := d.Attr("value")

		all = append(all, models.Programs{
			Day: value,
		})
	})

	checker := 0

	for i := 0; i < len(all); i++ {
		nURL := URL + "?date=" + all[i].Day

		resp, err := f.MakeRequest(nURL)
		if err != nil {
			log.Fatal(err)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		dateFormat, _ := time.Parse("20060102", all[i].Day)
		date := dateFormat.Format("02.01.2006")

		shouldStop := false

		doc.Find(
			"div[class='p-4 shadow rounded-lg flex items-center gap-6  transition-all duration-200 ']",
		).Each(func(j int, s *goquery.Selection) {
			if shouldStop {
				return
			}

			var start, finish string

			title := s.Find("h3.text-xl.font-semibold").Text()
			timeFull := s.Find("span.text-lg").Text()
			timeParts := strings.Split(timeFull, " - ")

			startTime, err := time.Parse("03:04 PM", timeParts[0])
			if err != nil {
				return
			}

			endTime, err := time.Parse("03:04 PM", timeParts[1])
			if err != nil {
				return
			}

			start = startTime.Format("03:04 PM")
			finish = endTime.Format("03:04 PM")

			startHalf := strings.Split(start, " ")[1]
			finishHalf := strings.Split(finish, " ")[1]

			if startHalf != finishHalf {
				checker++

				programs = append(programs, models.TimeTitle{
					Time: models.Time{
						Start:  convertTo24HourFormat(start),
						Finish: convertTo24HourFormat(finish),
					},
					Title:            title,
					Description:      "",
					Channel:          channelName,
					ChannelLogoURL:   channelLogoURL,
					AvailableArchive: availableArchive,
				})

				if checker == 2 {
					checker = 0
					shouldStop = true
					return
				}
			} else {
				programs = append(programs, models.TimeTitle{
					Time: models.Time{
						Start:  convertTo24HourFormat(start),
						Finish: convertTo24HourFormat(finish),
					},
					Title:            title,
					Description:      "",
					Channel:          channelName,
					ChannelLogoURL:   channelLogoURL,
					AvailableArchive: availableArchive,
				})
			}
		})

		all[i].Program = programs

		convertDate, err := f.ConvertCustomDate("02.01.2006", date)
		if err != nil {
			return nil, err
		}

		all[i].DayInt = convertDate
		all[i].Day = date

		programs = []models.TimeTitle{}
	}

	return all, nil
}

func convertTo24HourFormat(time12 string) string {
	splitTime := strings.Split(time12, " ")
	if len(splitTime) != 2 {
		return ""
	}

	timePart := splitTime[0]
	ampm := splitTime[1]

	splitTimeParts := strings.Split(timePart, ":")
	if len(splitTimeParts) != 2 {
		return ""
	}

	hourStr := splitTimeParts[0]
	minuteStr := splitTimeParts[1]

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return ""
	}

	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		return ""
	}

	if ampm == "AM" && hour == 12 {
		return fmt.Sprintf("00:%02d", minute)
	}

	if ampm == "PM" && hour != 12 {
		hour += 12
	}

	return fmt.Sprintf("%02d:%02d", hour, minute)
}
