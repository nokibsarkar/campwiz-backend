package importsources

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"os"
	"strconv"
)

type CSVListSource struct {
	CSVFilePath        string
	SubmissionIDColumn *string
	PageIDColumn       *string
	FileNameColumn     *string
	index              int // index is used to track the current row being processed
	data               any
}

func (t *ImporterServer) ImportFromCSV(ctx context.Context, req *models.ImportFromCSVRequest) (*models.ImportResponse, error) {
	source, err := NewCSVListSource(req.FilePath, req.SubmissionIdColumn, req.PageIdColumn, req.FileNameColumn)
	log.Printf("ImportFromCSV %v", req)
	if err != nil {
		log.Printf("Error creating CSVListSource: %s", err)
		return nil, err
	}
	if source == nil {
		log.Printf("CSVListSource is nil")
		return nil, errors.New("CSVListSource is nil")
	}
	go t.importFrom(ctx, source, req.TaskId, req.RoundId)
	return &models.ImportResponse{}, nil
}
func NewCSVListSource(csvFilePath string, submissionIDColumn string, pageIDColumn string, fileNameColumn string) (*CSVListSource, error) {
	c := &CSVListSource{CSVFilePath: csvFilePath}
	if submissionIDColumn != "" {
		c.SubmissionIDColumn = &submissionIDColumn
	}
	if pageIDColumn != "" {
		c.PageIDColumn = &pageIDColumn
	}
	if fileNameColumn != "" {
		c.FileNameColumn = &fileNameColumn
	}
	if c.data == nil {
		fp, err := os.Open(c.CSVFilePath)
		if err != nil {
			log.Printf("Error opening CSV file: %s", err)
			return nil, err
		}
		defer fp.Close()               //nolint:errcheck
		defer os.Remove(c.CSVFilePath) //nolint:errcheck
		// Delete The file after we are done
		// Read the CSV file
		reader := csv.NewReader(fp)
		if c.SubmissionIDColumn != nil {
			c.data, err = c.readTargetColumns(reader, *c.SubmissionIDColumn)
			if err != nil {
				return nil, err
			}
		} else if c.PageIDColumn != nil {
			c.data, err = c.readTargetColumns(reader, *c.PageIDColumn)
			if err != nil {
				return nil, err
			}
		}

		c.index = 0

	}
	return c, nil
}

// We would Read the CSV file
// If SubmissionIDcolumn is located, we are very happy and
// would directly import from our database
// But if PageIDColumn is located, we would need to get the
// data from wikimedia commons, but still not that bad
// becuase we can use 5000 images per request
// If FileNameColumn is located, we would need to get the
// data from wikimedia commons this time it is bad because
// it would be painfully slower because we have to search by
// file name which is not indexed
func (c *CSVListSource) ImportImageResults(ctx context.Context, currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	// Open the CSV file
	if c.SubmissionIDColumn != nil {
		return c.importUsingSubmissionId(ctx, currentRound, failedImageReason)
	} else if c.PageIDColumn != nil {
		return c.ImportImageResultsUsingPageId(ctx, currentRound, failedImageReason)
	}
	return nil, failedImageReason
}

func removeBOMFromString(s string) string {
	bom := []byte{0xEF, 0xBB, 0xBF}
	b := []byte(s)
	if bytes.HasPrefix(b, bom) {
		return string(b[len(bom):])
	}
	return s
}
func (c *CSVListSource) ImportImageResultsUsingPageId(ctx context.Context, currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	const batchSize = 15000
	pageIds, ok := c.data.([]string)
	if !ok {
		log.Printf("Data is not a slice of strings")
		(*failedImageReason)["*"] = "Data is not a slice of strings"
		return nil, failedImageReason
	}
	endIndex := min(c.index+batchSize, len(pageIds))
	log.Printf("Processing PageIDs from index %d to %d", c.index, endIndex)
	tempPageID := []uint64{}
	for c.index < endIndex {
		p := pageIds[c.index]
		c.index++
		p = removeBOMFromString(p)
		if p == "" {
			log.Printf("Skipping empty PageID")
			continue
		}
		d, err := strconv.Atoi(p)
		if err != nil {
			log.Printf("Error converting PageID to int: %s", err)
			continue
		}
		tempPageID = append(tempPageID, uint64(d))
		c.index++
	}
	result := []models.MediaResult{}
	if len(tempPageID) != 0 {
		q, close := repository.GetCommonsReplicaWithGen(ctx)
		defer close()
		data, err := q.CommonsSubmissionEntry.FetchSubmissionsFromCommonsDBByPageID(tempPageID, len(tempPageID))
		if err != nil {
			(*failedImageReason)["*"] = err.Error()
			log.Printf("Error importing images by PageID: %s", err)
			return nil, failedImageReason
		}
		for _, submission := range data {
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
	}
	return result, nil

}
func (c *CSVListSource) readTargetColumns(reader *csv.Reader, targetColumn string) ([]string, error) {
	headers, err := reader.Read()
	if err != nil {
		log.Printf("Error reading CSV file: %s", err)
		return nil, err
	}
	targetColumnIndex := -1
	for i, header := range headers {
		header = removeBOMFromString(header)
		if targetColumn == header {
			targetColumnIndex = i
			break
		}
	}
	if targetColumnIndex == -1 {
		log.Printf("Target column not found: %s", targetColumn)
		return nil, errors.New("target column not found: " + targetColumn)
	}
	var targetValues []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err == csv.ErrFieldCount {
			log.Printf("Error: Incorrect number of fields")
			return nil, errors.New("incorrect number of fields")
		} else if err != nil {
			log.Printf("Error reading CSV file: %s", err)
			return nil, err
		}
		targetValues = append(targetValues, record[targetColumnIndex])
	}
	return targetValues, nil
}
func (c *CSVListSource) importUsingSubmissionId(ctx context.Context, currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	submissionIds, ok := c.data.([]string)
	if !ok {
		log.Printf("Data is not a slice of strings")
		(*failedImageReason)["*"] = "Data is not a slice of strings"
		return nil, failedImageReason
	}
	q, close := repository.GetDBWithGen(ctx)
	defer close()
	Submission := q.Submission
	submissions, err := Submission.Select(Submission.ALL).Where(Submission.SubmissionID.In(submissionIds...)).Find()
	if err != nil {
		(*failedImageReason)["*"] = err.Error()
		log.Printf("Error getting submissions: %s", err)
		return nil, failedImageReason
	}
	if len(submissions) == 0 {
		(*failedImageReason)["*"] = "No submissions found"
		log.Printf("No submissions found")
		return nil, failedImageReason
	}

	results := make([]models.MediaResult, len(submissions))
	for i, submission := range submissions {
		results[i] = models.MediaResult{
			SubmissionID:     models.IDType(submission.SubmissionID),
			PageID:           submission.PageID,
			Name:             submission.Name,
			URL:              submission.URL,
			UploaderUsername: submission.Author,
			Height:           submission.Height,
			Width:            submission.Width,
			Size:             submission.Size,
			License:          submission.License,
			MediaType:        string(submission.MediaType),
			Duration:         submission.Duration,
			Description:      submission.Description,
			Resolution:       submission.Resolution,
			ThumbHeight:      &submission.ThumbHeight,
			ThumbWidth:       &submission.ThumbWidth,
			ThumbURL:         &submission.ThumbURL,
			CreditHTML:       submission.CreditHTML,
			SubmittedAt:      submission.SubmittedAt,
		}
	}
	return results, failedImageReason
}
