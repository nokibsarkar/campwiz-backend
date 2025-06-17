package models

import (
	"time"
)

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
type WikiMediaBaseResponse struct {
	Error *struct {
		Code    string `json:"code"`
		Info    string `json:"info"`
		Details string `json:"details"`
	} `json:"error"`
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

// {"edit":{"result":"Success","pageid":161628663,"title":"File:Cobbler repairing shoes in old workshop (Bazaar in Bitola, Macedonia, 2025).jpg","contentmodel":"wikitext","oldrevid":1044733089,"newrevid":1044733280,"newtimestamp":"2025-06-16T22:47:19Z"}}
type EditResponse struct {
	Result   string  `json:"result"`
	PageID   uint64  `json:"pageid"`
	Title    string  `json:"title"`
	Content  string  `json:"contentmodel"`
	OldRev   uint64  `json:"oldrevid"`
	NewRev   uint64  `json:"newrevid"`
	NewTS    string  `json:"newtimestamp"`
	NoChange *string `json:"nochange,omitempty"` // Optional, may not be present if there was no change
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
type PageContent string

// {"revid":1042731719,"parentid":1038162825,"user":"UnpetitproleX","timestamp":"2025-06-11T13:36:16Z","slots":{"main":{"contentmodel":"wikitext","contentformat":"text/x-wiki","*":
type Slots struct {
	Main struct {
		ContentModel  string      `json:"contentmodel"`
		ContentFormat string      `json:"contentformat"`
		Content       PageContent `json:"*"`
	} `json:"main"`
}

type Revision struct {
	Revid     uint64                `json:"revid"`
	ParentID  uint64                `json:"parentid"`
	User      WikimediaUsernameType `json:"user"`
	Timestamp time.Time             `json:"timestamp"`
	Comment   string                `json:"comment"`
	Slots     *Slots                `json:"slots,omitempty"` // Optional, may not be present in all revisions
	Page      Page                  `json:"-"`
}

type RevisionPage struct {
	Page
	Revisions []Revision `json:"revisions"`
}

const COMMONS_CATEGORY_NAMESPACE = "Category"

var CATEGORY_PATTERN_WITH_SORTKEY = `\[\[ *?[Cc]ategory *?:(?P<categoryName>[^\]\|]+?)(?:\|(?P<sortKey>[^\]]+?))?\]\]`

type TokenType string

const (
	TokenTypeCategory  TokenType = "category"
	TokenTypeReference TokenType = "reference"
	TokenTypeOther     TokenType = "other"
)

type PageCategory struct {
	// The name of the category, Without the namespace
	Name    string `json:"name"` // The name of the category, without namespace
	SortKey string `json:"-"`    // Sort key is optional, can be empty
	// Whether the category is fixed and cannot be removed
	Fixed bool `json:"fixed"` // Whether the category is fixed and cannot be removed
}
type SubmissionWithCategoryList struct {
	Submission
	Categories []PageCategory `json:"categories"` // List of categories associated with the submission
}
type BaseEditResponse struct {
	WikiMediaBaseResponse
	Edit EditResponse `json:"edit"`
}
