package models

import "time"

type Programs struct {
	Day     string
	DayInt  time.Time
	Program []TimeTitle
	Today   bool
}

type TimeTitle struct {
	Time             Time
	Title            string
	Description      string
	Channel          string
	ChannelLogoURL   string
	AvailableArchive int
}

type Time struct {
	Start     string
	Finish    string
	StartISO  string
	FinishISO string
}

type SmartPrograms struct {
	Title  string
	Start  string
	Finish string
}
