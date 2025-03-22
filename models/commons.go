package models

import "time"

type MediaResult struct {
	PageID           uint64 `json:"pageid"`
	SubmissionID     IDType `json:"-"`
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
	ThumbURL         *string
	ThumbWidth       *uint64
	ThumbHeight      *uint64
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
type Thumbnail struct {
	ThumbURL    string `json:"thumburl"`
	ThumbWidth  uint64 `json:"thumbwidth"`
	ThumbHeight uint64 `json:"thumbheight"`
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
		*Thumbnail
	} `json:"imageinfo"`
}
type Page struct {
	PageID int    `json:"pageid"`
	Ns     int    `json:"ns"`
	Title  string `json:"title"`
}
type ImageInfoPage struct {
	Page
	ImageInfo
}
