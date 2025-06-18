package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"strings"
	"time"

	"github.com/st3fan/html2text"
)

const COMMONS_API = "https://commons.wikimedia.org/w/api.php"
const CAMPWIZ_USER_AGENT = "Campwiz/1.0 (https://campwiz.nokib.com; +https://github.com/nokibsarkar/campwiz)"

var COMMONS_AUDIO_THUMB = "https://commons.wikimedia.org/w/resources/assets/file-type-icons/fileicon-ogg.png"
var COMMONS_THUMB_WIDTH uint64 = 400
var COMMONS_THUMB_HEIGHT uint64 = 400

// This Repository would be used to communicate with wikimedia commons
type CommonsRepository struct {
	endpoint    string
	accessToken string
	cl          *http.Client
	csrf        string // CSRF token for editing
}

// NewCommonsRepository returns a new instance of CommonsRepository
func NewCommonsRepository(cl *http.Client) *CommonsRepository {
	accessToken := ""
	if cl == nil {
		cl = &http.Client{}
		accessToken = consts.Config.Auth.AccessToken
	}
	return &CommonsRepository{
		endpoint:    COMMONS_API,
		accessToken: accessToken,
		cl:          cl,
	}
}
func (c *CommonsRepository) Get(values url.Values) (_ io.ReadCloser, err error) {
	// Get values from commons
	url := fmt.Sprintf("%s?%s", c.endpoint, values.Encode())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.accessToken != "" {
		// Set the access token in the header
		// This is used for authenticated requests
		// If the access token is not set, it will be a public request
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}
	// Set the user agent to commons
	req.Header.Set("User-Agent", CAMPWIZ_USER_AGENT)
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
func (c *CommonsRepository) POST(values url.Values, body map[string]string) (_ io.ReadCloser, err error) {

	buf := bytes.NewBuffer(nil)

	url := fmt.Sprintf("%s?%s", c.endpoint, values.Encode())

	mp := multipart.NewWriter(buf)
	{
		defer mp.Close() //nolint:errcheck
		for k, v := range body {
			if err := mp.WriteField(k, v); err != nil {
				return nil, err
			}
		}

	}

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return nil, err
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+mp.Boundary())
	// Set the user agent to commons
	req.Header.Set("User-Agent", CAMPWIZ_USER_AGENT)
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close() //nolint:errcheck
		bodyBytes, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("failed to post to commons: %s, body: %s", resp.Status, string(bodyBytes))
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
		"iiprop":    {"timestamp|user|url|size|mediatype|dimensions|canonicaltitle"},
		"gcmlimit":  {"max"},
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
		// log.Println("Image info: ", info.Title)
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
		} else {
			thumbURL, thumbWidth, thumbHeight := c.GetImageThumbFromURL(info.URL, float32(info.Width)/float32(info.Height), COMMONS_THUMB_WIDTH)
			img.ThumbURL = &thumbURL
			img.ThumbWidth = &thumbWidth
			img.ThumbHeight = &thumbHeight
		}
		result = append(result, img)
	}
	return result, map[string]string{}
}
func (c *CommonsRepository) GetImagesFromCommonsCategories2(ctx context.Context, category string, lastPageID uint64, round *models.Round, startDate time.Time, endDate time.Time) (result []models.MediaResult, currentfailedImages map[string]string, lastPageIDOut uint64) {
	q, close := GetCommonsReplicaWithGen(ctx)
	defer close()
	log.Printf("1 Getting images from commons category: %s", category)
	result = []models.MediaResult{}
	currentfailedImages = map[string]string{}
	const batchSize = 15000
	startDateInt := models.Date2Int(startDate)
	endDateInt := models.Date2Int(endDate)
	allowedtypes := []string{}
	for _, mediatype := range round.AllowedMediaTypes {
		allowedtypes = append(allowedtypes, string(mediatype))
	}

	lastCount := batchSize
	for lastCount == batchSize {
		log.Println("Getting images from commons category: ", category)
		ssubmissionChunk, err := q.CommonsSubmissionEntry.FetchSubmissionsFromCommonsDBByCategory(category, lastPageID, startDateInt, endDateInt, batchSize, allowedtypes)
		if err != nil {
			log.Println("Error: ", err)
			return
		}
		log.Println("Submissions: ", len(ssubmissionChunk))
		lastCount = len(ssubmissionChunk)
		for _, submission := range ssubmissionChunk {
			if submission.PageID > lastPageID {
				lastPageID = submission.PageID
				lastPageIDOut = lastPageID
			} else {
				log.Println("Skipping submission: ", submission.PageID)
			}
			thumbURL, thumbWidth, thumbHeight := submission.GetThumbURL()
			result = append(result, models.MediaResult{
				PageID:           submission.PageID,
				Name:             submission.PageTitle,
				URL:              submission.GetURL(),
				UploaderUsername: models.WikimediaUsernameType(submission.UserName),
				SubmittedAt:      submission.GetSubmittedAt(),
				Height:           submission.FrHeight,
				Width:            submission.FrWidth,
				Size:             submission.FrSize,
				MediaType:        submission.FtMediaType,
				Resolution:       submission.FrWidth * submission.FrHeight,
				ThumbURL:         &thumbURL,
				ThumbWidth:       &thumbWidth,
				ThumbHeight:      &thumbHeight,
			})
		}
		log.Println("Last Page ID: ", lastPageID)
	}
	if len(result) == 0 {
		log.Println("No images found in commons category: ", category)
		lastPageIDOut = 0
		return
	}
	lastPageIDOut = lastPageID
	// Get images from commons category
	// Create batch from commons category
	return
}
func (c *CommonsRepository) GetImageThumbFromURL(fileURL string, aspectRatio float32, targetWidth uint64) (string, uint64, uint64) {
	// file name is the last part of the URL
	fileNameWithoutPrefix := fileURL[strings.LastIndex(fileURL, "/")+1:]
	// extract the file extension
	// extension := strings.ToLower(fileNameWithoutPrefix[strings.LastIndex(fileNameWithoutPrefix, "."):])
	// thumbSuffix := fileNameWithoutPrefix
	// switch extension {
	// case ".jpg", ".jpeg", ".png", ".webp":
	// 	// For image files, we can set extension as is
	// 	thumbSuffix = fileNameWithoutPrefix
	// default:
	// 	// For SVG and GIF files, we can use a PNG thumbnail
	// 	thumbSuffix = fileNameWithoutPrefix + ".png"
	// }
	// Calculate the width and height of the thumbnail
	thumbWidth := targetWidth
	// aspectRatio is the ratio of width to height
	// thumbHeight = thumbWidth / aspectRatio
	thumbHeight := uint64(float32(targetWidth) / aspectRatio)
	thumbURL := fmt.Sprintf("https://commons.wikimedia.org/w/thumb.php?f=%s&width=%d&height=%d", fileNameWithoutPrefix, thumbWidth, thumbHeight)
	return thumbURL, thumbWidth, thumbHeight
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
func (c *CommonsRepository) GetImagesDescriptionFromIPageIDs(pageids []uint64) []models.MediaResult {
	// Get images from commons category
	// Create batch from commons category
	start := 0
	total := len(pageids)
	batchSize := 500

	result := []models.MediaResult{}
	for start < total {
		end := min(start+batchSize, total)
		batch := []string{}
		for _, pageid := range pageids[start:end] {
			batch = append(batch, fmt.Sprintf("%d", pageid))
		}
		paginator := NewPaginator[models.ImageInfoPage](c)
		params := url.Values{
			"action":              {"query"},
			"format":              {"json"},
			"prop":                {"imageinfo"},
			"pageids":             {strings.Join(batch, "|")},
			"iiprop":              {"extmetadata"},
			"iilimit":             {"1"},
			"limit":               {"max"},
			"iiextmetadatafilter": {"ImageDescription"},
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
			extmetadata := info.ExtMetadata
			if extmetadata == nil {
				log.Println("No extmetadata found. Skipping")
				continue
			}
			img := models.MediaResult{
				PageID:      uint64(image.PageID),
				Description: extmetadata.GetImageDescription(),
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
			// if user.Registered.IsZero() {
			// 	log.Printf("No registration date found for %s. Skipping", user.Name)
			// }
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

type tokensResponse struct {
	Tokens struct {
		CSRF *string `json:"csrftoken"`
	} `json:"tokens"`
}

func (c *CommonsRepository) GetCsrfToken() (string, error) {
	if c.csrf != "" {
		return c.csrf, nil
	}
	data := url.Values{
		"action": {"query"},
		"meta":   {"tokens"},
		"format": {"json"},
	}
	resp, err := c.Get(data)
	if err != nil {
		return "", err
	}
	defer resp.Close() //nolint:errcheck
	decoder := json.NewDecoder(resp)
	var response BaseQueryResponse[tokensResponse, map[string]string]
	if err := decoder.Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if response.Error != nil {
		return "", fmt.Errorf("error from commons API: %s - %s", response.Error.Code, response.Error.Info)
	}
	if response.Query.Tokens.CSRF == nil {
		return "", errors.New("noCSRF")
	}
	c.csrf = *response.Query.Tokens.CSRF
	return *response.Query.Tokens.CSRF, nil
}
func (c *CommonsRepository) GetEditToken() (string, error) {
	// Get edit token
	return c.GetCsrfToken()
}
func (c *CommonsRepository) GetUserInfo() {
	// Get user info
}

type BaseQueryResponse[QueryType any, ContinueType map[string]string] struct {
	models.WikiMediaBaseResponse
	BatchComplete *string       `json:"batchcomplete"`
	Next          *ContinueType `json:"continue"`
	Query         QueryType     `json:"query"`
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

func (c *CommonsRepository) GetLatestPageRevisionByPageID(ctx context.Context, pageID uint64) (*models.Revision, error) {
	qs := url.Values{
		"action":   {"query"},
		"format":   {"json"},
		"prop":     {"revisions"},
		"pageids":  {fmt.Sprintf("%d", pageID)},
		"rvprop":   {"ids|timestamp|user|comment|content"},
		"rvslots":  {"main"},
		"rvlimit":  {"1"},
		"continue": {""},
	}
	resp, err := c.Get(qs)
	if err != nil {
		return nil, err
	}
	defer resp.Close() //nolint:errcheck
	decoder := json.NewDecoder(resp)
	var response PageQueryResponse[models.RevisionPage]
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("error from commons API: %s - %s", response.Error.Code, response.Error.Info)
	}
	if len(response.Query.Pages) == 0 {
		return nil, fmt.Errorf("no pages found for page ID %d", pageID)
	}
	page, exists := response.Query.Pages[fmt.Sprintf("%d", pageID)]
	if !exists {
		return nil, fmt.Errorf("no page found for page ID %d", pageID)
	}
	if len(page.Revisions) == 0 {
		return nil, fmt.Errorf("no revisions found for page ID %d", pageID)
	}
	revision := &page.Revisions[0]
	revision.Page = page.Page
	return revision, nil
}
func (c *CommonsRepository) EditPageContent(ctx context.Context, pageID uint64, content string, summary string) error {
	csrfToken, err := c.GetCsrfToken()
	if err != nil {
		return fmt.Errorf("failed to get CSRF token: %w", err)
	}
	data := map[string]string{
		"pageid":  fmt.Sprintf("%d", pageID),
		"text":    content,
		"summary": summary,
		"token":   csrfToken,
	}
	qs := url.Values{
		"action": {"edit"},
		"format": {"json"},
		"assert": {"user"}, // Ensure the user is authenticated
		"pageid": {fmt.Sprintf("%d", pageID)},
	}
	resp, err := c.POST(qs, data)
	if err != nil {
		return err
	}
	defer resp.Close() //nolint:errcheck
	response := &models.BaseEditResponse{}
	if err := json.NewDecoder(resp).Decode(response); err != nil {
		return fmt.Errorf("decode-err-%w", err)
	}
	if response.Error != nil {
		log.Printf("Error editing page %d: %s - %s", pageID, response.Error.Code, response.Error.Info)
		return fmt.Errorf("wikimediaAPI.%s", response.Error.Code)
	}
	if response.Edit.NoChange != nil {
		log.Printf("No change made to page %d: %s", pageID, response.Edit.Title)
		return nil // No change made, but not an error
	}
	log.Printf("Successfully edited page %d with summary: %s", pageID, summary)
	return nil
}
