package models

import "time"

type ImageResult struct {
	ID               uint64 `json:"pageid"`
	Name             string `json:"title"`
	URL              string
	SubmittedAt      time.Time
	UploaderUsername WikimediaUsernameType
	Height           uint64
	Width            uint64
	Size             uint64
	MediaType        string
	Duration         uint64
	License          string
	Description      string
	CreditHTML       string
	Resolution       uint64
}
type GContinue struct {
	Gcmcontinue string `json:"gcmcontinue"`
}

type WikimediaUser struct {
	Name       WikimediaUsernameType `json:"name"`
	Registered time.Time             `json:"registration"`
	CentralIds *struct {
		CentralAuth uint64 `json:"CentralAuth"`
	} `json:"centralids"`
}
type UserList struct {
	Users map[string]WikimediaUser `json:"users"`
}

type KeyValue struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type ImageInfo struct {
	Info []struct {
		Timestamp      time.Time             `json:"timestamp"`
		User           WikimediaUsernameType `json:"user"`
		Size           uint64                `json:"size"`
		Width          uint64                `json:"width"`
		Height         uint64                `json:"height"`
		Title          string                `json:"canonicaltitle"`
		URL            string                `json:"url"`
		DescriptionURL string                `json:"descriptionurl"`
		MediaType      string                `json:"mediatype"`
		Duration       float64               `json:"duration"`
		ExtMetadata    *ExtMetadata          `json:"extmetadata"`
	} `json:"imageinfo"`
}
type Page struct {
	Pageid int    `json:"pageid"`
	Ns     int    `json:"ns"`
	Title  string `json:"title"`
}
type ImageInfoPage struct {
	Page
	ImageInfo
}
