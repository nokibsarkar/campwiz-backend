package repository

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"strings"

	"github.com/st3fan/html2text"
)

const COMMONS_API = "http://commons.wikimedia.org/w/api.php"

var COMMONS_AUDIO_THUMB = "https://commons.wikimedia.org/w/resources/assets/file-type-icons/fileicon-ogg.png"
var COMMONS_THUMB_WIDTH uint64 = 512
var COMMONS_THUMB_HEIGHT uint64 = 512

// This Repository would be used to communicate with wikimedia commons
type CommonsRepository struct {
	endpoint    string
	accessToken string
	cl          *http.Client
}

func (c *CommonsRepository) Get(values url.Values) (_ io.ReadCloser, err error) {
	// Get values from commons
	url := fmt.Sprintf("%s?%s", c.endpoint, values.Encode())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// returns images from commons categories
func (c *CommonsRepository) GetImagesFromCommonsCategories(category string) ([]models.MediaResult, map[string]string) {
	// Get images from commons category
	// Create batch from commons category
	log.Println("Getting images from commons category: ", category)
	paginator := NewPaginator[models.ImageInfoPage](c)
	params := url.Values{
		"action":    {"query"},
		"format":    {"json"},
		"prop":      {"imageinfo"},
		"generator": {"categorymembers"},
		"gcmtitle":  {category},
		"gcmtype":   {"file"},
		"iiprop":    {"timestamp|user|url|size|mediatype|dimensions|extmetadata|canonicaltitle"},
		"limit":     {"max"},
		// "iiurlwidth":          {"640"},
		// "iiurlheight":         {"480"},
		"iiextmetadatafilter": {"License|ImageDescription|Credit|Artist|LicenseShortName|UsageTerms|AttributionRequired|Copyrighted"},
	}
	images, err := paginator.Query(params)
	if err != nil {
		log.Println("Error: ", err)
		return nil, nil
	}
	result := []models.MediaResult{}
	for image := range images {
		// Append images to result
		if image == nil {
			break
		}
		if len(image.Info) == 0 {
			log.Println("No image info found. Skipping")
			continue
		}
		info := image.Info[0]
		img := models.MediaResult{
			PageID:           uint64(image.PageID),
			Name:             image.Title,
			URL:              info.URL,
			UploaderUsername: info.User,
			SubmittedAt:      info.Timestamp,
			Height:           info.Height,
			Width:            info.Width,
			Size:             info.Size,
			MediaType:        info.MediaType,
			Duration:         uint64(info.Duration * 1e3), // Convert to milliseconds
			Resolution:       info.Width * info.Height,
		}
		if info.ExtMetadata != nil {
			img.License = html2text.HTML2Text(info.ExtMetadata.GetLicense())
			img.Description = html2text.HTML2Text(info.ExtMetadata.GetImageDescription())
			img.CreditHTML = info.ExtMetadata.GetCredit()
		}
		if info.MediaType == string(models.MediaTypeAudio) {
			img.ThumbURL = &COMMONS_AUDIO_THUMB
			img.ThumbWidth = &COMMONS_THUMB_WIDTH
			img.ThumbHeight = &COMMONS_THUMB_HEIGHT
		}
		result = append(result, img)
	}
	return result, map[string]string{}
}

// returns images from commons categories
func (c *CommonsRepository) GetImagesThubsFromCommonsCategories(category string) ([]models.MediaResult, map[string]string) {
	// Get images from commons category
	// Create batch from commons category
	log.Println("Getting images from commons category: ", category)
	paginator := NewPaginator[models.ImageInfoPage](c)
	params := url.Values{
		"action":      {"query"},
		"format":      {"json"},
		"prop":        {"imageinfo"},
		"generator":   {"categorymembers"},
		"gcmtitle":    {category},
		"gcmtype":     {"file"},
		"iiprop":      {"url"},
		"limit":       {"50"},
		"iiurlwidth":  {"640"},
		"iiurlheight": {"480"},
	}
	images, err := paginator.Query(params)
	if err != nil {
		log.Println("Error: ", err)
		return nil, nil
	}
	result := []models.MediaResult{}
	for image := range images {
		// Append images to result
		if image == nil {
			break
		}
		if len(image.Info) == 0 {
			log.Println("No image info found. Skipping")
			continue
		}
		info := image.Info[0]
		img := models.MediaResult{
			PageID:      uint64(image.PageID),
			Name:        image.Title,
			URL:         info.URL,
			ThumbURL:    &info.ThumbURL,
			ThumbHeight: &info.ThumbHeight,
			ThumbWidth:  &info.Width,
		}
		result = append(result, img)
	}
	log.Println("Found images: ", len(result))
	return result, map[string]string{}
}

// returns images from commons categories
func (c *CommonsRepository) GetImagesThumbsFromIPageIDs(pageids []uint64) []models.MediaResult {
	// Get images from commons category
	// Create batch from commons category
	start := 0
	total := len(pageids)

	result := []models.MediaResult{}
	for start < total {
		end := min(start+50, total)
		batch := []string{}
		for _, pageid := range pageids[start:end] {
			batch = append(batch, fmt.Sprintf("%d", pageid))
		}
		paginator := NewPaginator[models.ImageInfoPage](c)
		params := url.Values{
			"action":      {"query"},
			"format":      {"json"},
			"prop":        {"imageinfo"},
			"pageids":     {strings.Join(batch, "|")},
			"iiprop":      {"url"},
			"limit":       {"50"},
			"iilimit":     {"1"},
			"iiurlwidth":  {"640"},
			"iiurlheight": {"480"},
		}
		log.Println("Getting images from commons pageids: ", strings.Join(batch, "|"))
		images, err := paginator.Query(params)
		if err != nil {
			log.Println("Error: ", err)
			return nil
		}
		for image := range images {
			// Append images to result
			if image == nil {
				break
			}
			if len(image.Info) == 0 {
				log.Println("No image info found. Skipping")
				continue
			}
			info := image.Info[0]
			img := models.MediaResult{
				PageID:      uint64(image.PageID),
				ThumbURL:    &info.ThumbURL,
				ThumbHeight: &info.ThumbHeight,
				ThumbWidth:  &info.Width,
			}
			result = append(result, img)
		}
		start = end
		log.Println("Found images: ", len(result))
	}
	return result
}

// returns images from commons categories
func (c *CommonsRepository) GeUsersFromUsernames(usernames []models.WikimediaUsernameType) ([]models.WikimediaUser, error) {
	// Get images from commons category
	// Create batch from commons category
	paginator := NewPaginator[models.WikimediaUser](c)
	batchSize := 40
	batchCount := len(usernames) / batchSize
	if len(usernames)%batchSize != 0 {
		batchCount++
	}
	result := []models.WikimediaUser{}
	for i := range batchCount {
		start := i * batchSize
		end := min((i+1)*batchSize, len(usernames))
		batch := []string{}
		for _, username := range usernames[start:end] {
			batch = append(batch, string(username))
		}
		params := url.Values{
			"action":  {"query"},
			"format":  {"json"},
			"list":    {"users"},
			"ususers": {strings.Join(batch, "|")},
			"usprop":  {"centralids|registration"},
			"limit":   {"max"},
		}
		users, err := paginator.UserList(params)
		if err != nil {
			log.Println("Error: ", err)
			return nil, nil
		}
		for user := range users {
			// Append images to result
			if user == nil {
				break
			}
			if user.Registered.IsZero() {
				log.Println("No registration date found. Skipping")
				continue
			}
			result = append(result, *user)
		}
	}
	return result, nil
}
func (c *CommonsRepository) GetImageDetails() {
	// Get image details
}
func (c *CommonsRepository) GetImageMetadata() {
	// Get image metadata
}
func (c *CommonsRepository) GetImageCategories() {
	// Get image categories
}
func (c *CommonsRepository) GetCsrfToken() {
	// Get csrf token
}
func (c *CommonsRepository) GetEditToken() {
	// Get edit token
}
func (c *CommonsRepository) GetUserInfo() {
	// Get user info
}

type BaseQueryResponse[QueryType any, ContinueType map[string]string] struct {
	BatchComplete string        `json:"batchcomplete"`
	Next          *ContinueType `json:"continue"`
	Query         QueryType     `json:"query"`
	Error         *struct {
		Code    string `json:"code"`
		Info    string `json:"info"`
		Details string `json:"details"`
	} `json:"error"`
}
type PageQueryResponse[PageType any] = BaseQueryResponse[struct {
	Normalized []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"normalized"`
	Pages map[string]PageType `json:"pages"`
}, map[string]string]

type UserListQueryResponse = BaseQueryResponse[struct {
	Users []models.WikimediaUser `json:"users"`
}, map[string]string]

// NewCommonsRepository returns a new instance of CommonsRepository
func NewCommonsRepository() *CommonsRepository {
	return &CommonsRepository{
		endpoint:    COMMONS_API,
		accessToken: consts.Config.Auth.AccessToken,
		cl:          &http.Client{},
	}
}
