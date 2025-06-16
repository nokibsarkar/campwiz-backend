package models

import (
	"fmt"
	"log"
	"regexp"
	"strings"
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
type Token struct {
	Start   int
	End     int
	Content string
	Entity  any // This can be used to store additional information about the token, like a category name or sort key
	Type    TokenType
}
type CategoryMap map[string]Token

func (c *CategoryMap) GetContent() *PageContent {
	return (*c)[" "].Entity.(*PageContent)
}
func (c *CategoryMap) GetTokens() []Token {
	return (*c)[""].Entity.([]Token)
}
func (c *CategoryMap) Add(category string) (*CategoryMap, bool) {
	pageContent := c.GetContent()
	if pageContent == nil {
		return c, false
	}
	if category == "" {
		return c, false
	}
	if _, ok := (*c)[category]; ok {
		return c, false
	}
	// add the category at the end of the content
	tokens := c.GetTokens()
	formattedCategory := strings.TrimSpace(category)
	formattedCategory = fmt.Sprintf("[[Category:%s]]", formattedCategory)
	(*pageContent) = PageContent(
		string(*pageContent) + "\n" + formattedCategory,
	)
	newToken := Token{
		Start:   len(string(*pageContent)) - len(formattedCategory),
		End:     len(string(*pageContent)),
		Content: formattedCategory,
		Type:    TokenTypeCategory,
		Entity: PageCategory{
			Name: category,
		},
	}
	(*c)[category] = newToken
	// update the token list
	tokens = append(tokens, newToken)
	(*c)[""] = Token{
		Start:   0,
		End:     len(string(*pageContent)),
		Content: "TokenListRef$",
		Type:    TokenTypeOther,
		Entity:  tokens,
	}
	(*c)[" "] = Token{
		Start:   0,
		End:     0,
		Content: "PageContentRef$",
		Type:    TokenTypeOther,
		Entity:  pageContent,
	}
	log.Printf("Added category: %s, new content length: %d", category, len(string(*pageContent)))
	log.Printf("New token: %+v", newToken)
	log.Printf("New category map: %+v", *c)
	return c, true
}
func (con *PageContent) GetCategoryMappingFromTokenList(tokens []Token) *CategoryMap {
	if con == nil || tokens == nil {
		return nil
	}
	categoryMap := make(CategoryMap)
	categoryMap[" "] = Token{
		Start:   0,
		End:     0,
		Content: "PageContentRef$",
		Type:    TokenTypeOther,
		Entity:  con,
	}
	categoryMap[""] = Token{
		Start:   0,
		End:     0,
		Content: "TokenListRef$",
		Type:    TokenTypeOther,
		Entity:  tokens,
	}
	for _, token := range tokens {
		if token.Type == TokenTypeCategory {
			if category, ok := token.Entity.(PageCategory); ok {
				categoryMap[category.Name] = token
			}
		}
	}
	return &categoryMap
}

func (con *PageContent) SplitIntoTokens() []Token {
	if con == nil {
		return nil
	}
	content := string(*con)
	tokens := make([]Token, 0)
	patt := regexp.MustCompile(`\[\[ *?[Cc]ategory *?:(?P<categoryName>[^\]\|]+?)(?:\|(?P<sortKey>[^\]]+?))?\]\]`)
	if patt == nil {
		log.Println("Failed to compile regex pattern for categories")
		return nil
	}
	matches := patt.FindAllStringSubmatchIndex(content, -1)
	if matches == nil {
		return nil
	}
	lastEnd := 0
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		startPosition := match[0]
		endPosition := match[1]
		if startPosition > lastEnd {
			// add a text token for the content before the category
			tokens = append(tokens, Token{
				Start:   lastEnd,
				End:     startPosition,
				Content: content[lastEnd:startPosition],
				Type:    TokenTypeOther,
			})
		}
		name := content[match[2]:match[3]]
		sortKey := ""
		if len(match) > 4 && match[4] != -1 {
			sortKey = content[match[4]:match[5]]
		}
		tokens = append(tokens, Token{
			Start:   startPosition,
			End:     endPosition,
			Content: content[startPosition:endPosition],
			Type:    TokenTypeCategory,
			Entity: PageCategory{
				Name:    strings.TrimSpace(name),
				SortKey: strings.TrimSpace(sortKey),
			},
		})
		lastEnd = endPosition
	}
	return tokens
}
func (con *PageContent) GetCategoriesWithoutNamepace() []string {
	if con == nil {
		return nil

	}
	patt := regexp.MustCompile(CATEGORY_PATTERN_WITH_SORTKEY)
	if patt == nil {
		log.Println("Failed to compile regex pattern for categories")
		return nil
	}
	content := string(*con)
	matches := patt.FindAllStringSubmatch(content, -1)
	if matches == nil {
		return nil
	}
	categories := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		category := strings.TrimSpace(match[1])
		if category == "" {
			continue
		}
		categories = append(categories, category)
	}
	return categories
}
