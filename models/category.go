package models

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

// This is a representative category model for the submission category.
// It has many to one relationship with the submission model.
type CategoryWithWriteableFields struct {
	CategoryName string  `json:"categoryName" gorm:"not null;primaryKey;uniqueIndex:idx_submission_category;index"`
	SubmissionID IDType  `json:"submissionId" gorm:"not null;primaryKey;uniqueIndex:idx_submission_category"` // The submission this category belongs to
	AddedByID    *IDType `json:"addedById" gorm:"null;index;uniqueIndex:idx_submission_category"`             // The user who added this category
}
type Category struct {
	CategoryWithWriteableFields
	CreatedAt  *time.Time  `json:"createdAt" gorm:"<-:create"`                                                    // The time the category was created
	AddedBy    *User       `json:"-" gorm:"foreignKey:AddedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`   // The user who added this category
	Submission *Submission `json:"-" gorm:"foreignKey:SubmissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // The submission this category belongs to
	gorm.DeletedAt
}

// This would be as response after a user submits categories for a submission.
type CategoryResponse struct {
	// PageTitle is the title of the page where the categories were added or removed
	PageTitle string   `json:"pageTitle"` //
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Executed  bool     `json:"executed"` // Whether the categories were added or removed successfully
}
type TextToken struct {
	Start   int
	End     int
	Content string
	Entity  any // This can be used to store additional information about the token, like a category name or sort key
	Type    TokenType
}
type CategoryMap map[string]TextToken

func ToCanonicalName(name string) string {
	// This function is used to get the canonical name of a category
	// It removes the namespace and returns the name in lowercase
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = norm.NFC.String(name) // Normalize the string to NFC form

	name = strings.ToUpper(name[:1]) + name[1:]
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
func FromCanonicalName(name string) string {
	// This function is used to get the original name of a category
	// It replaces underscores with spaces
	if name == "" {
		return ""
	}
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "_", " ")
	return name
}
func (c *CategoryMap) GetContent() *PageContent {
	return (*c)[" "].Entity.(*PageContent)
}
func (c *CategoryMap) GetTokens() []TextToken {
	return (*c)[""].Entity.([]TextToken)
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
	newToken := TextToken{
		Start:   len(string(*pageContent)) - len(formattedCategory),
		End:     len(string(*pageContent)),
		Content: formattedCategory,
		Type:    TokenTypeCategory,
		Entity: PageCategory{
			Name: category,
		},
	}
	(*c)[ToCanonicalName(category)] = newToken
	// update the token list
	tokens = append(tokens, newToken)
	(*c)[""] = TextToken{
		Start:   0,
		End:     len(string(*pageContent)),
		Content: "TokenListRef$",
		Type:    TokenTypeOther,
		Entity:  tokens,
	}
	(*c)[" "] = TextToken{
		Start:   0,
		End:     0,
		Content: "PageContentRef$",
		Type:    TokenTypeOther,
		Entity:  pageContent,
	}
	return c, true
}
func (con *PageContent) GetCategoryMappingFromTokenList(tokens []TextToken) *CategoryMap {
	if con == nil || tokens == nil {
		return nil
	}
	categoryMap := make(CategoryMap)
	categoryMap[" "] = TextToken{
		Start:   0,
		End:     0,
		Content: "PageContentRef$",
		Type:    TokenTypeOther,
		Entity:  con,
	}
	categoryMap[""] = TextToken{
		Start:   0,
		End:     0,
		Content: "TokenListRef$",
		Type:    TokenTypeOther,
		Entity:  tokens,
	}
	for _, token := range tokens {
		if token.Type == TokenTypeCategory {
			if category, ok := token.Entity.(PageCategory); ok {
				categoryMap[ToCanonicalName(category.Name)] = token
			}
		}
	}
	return &categoryMap
}

func (con *PageContent) SplitIntoTokens() []TextToken {
	if con == nil {
		return nil
	}
	content := string(*con)
	tokens := make([]TextToken, 0)
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
			tokens = append(tokens, TextToken{
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
		tokens = append(tokens, TextToken{
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
