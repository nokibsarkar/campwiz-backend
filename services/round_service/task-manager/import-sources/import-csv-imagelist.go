package importsources

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	"os"
	"strings"
)

type CSVListSource struct {
	CSVFilePath        string
	SubmissionIDColumn *string
	PageIDColumn       *string
	FileNameColumn     *string
}

func (t *ImporterServer) ImportFromCSV(ctx context.Context, req *models.ImportFromCSVRequest) (*models.ImportResponse, error) {
	source := NewCSVListSource(req.FilePath, req.SubmissionIdColumn, req.PageIdColumn, req.FileNameColumn)
	log.Printf("ImportFromCSV %v", req)
	go t.importFrom(ctx, source, req.TaskId, req.RoundId)
	return &models.ImportResponse{}, nil
}
func NewCSVListSource(csvFilePath string, submissionIDColumn string, pageIDColumn string, fileNameColumn string) *CSVListSource {
	data := &CSVListSource{CSVFilePath: csvFilePath}
	if submissionIDColumn != "" {
		data.SubmissionIDColumn = &submissionIDColumn
	}
	if pageIDColumn != "" {
		data.PageIDColumn = &pageIDColumn
	}
	if fileNameColumn != "" {
		data.FileNameColumn = &fileNameColumn
	}
	return data
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
	fp, err := os.Open(c.CSVFilePath)
	if err != nil {
		(*failedImageReason)["*"] = err.Error()
		log.Printf("Error opening CSV file: %s", err)
		return nil, failedImageReason
	}
	defer fp.Close() //nolint:errcheck
	// Delete The file after we are done
	defer os.Remove(c.CSVFilePath) //nolint:errcheck
	// Read the CSV file
	reader := csv.NewReader(fp)
	if c.SubmissionIDColumn != nil {
		return c.importUsingSubmissionId(ctx, reader, failedImageReason)
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
func (c *CSVListSource) importUsingSubmissionId(ctx context.Context, reader *csv.Reader, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	headers, err := reader.Read()
	if err != nil {
		(*failedImageReason)["*"] = err.Error()
		log.Printf("Error reading CSV file: %s", err)
		return nil, failedImageReason
	}
	submissionIDColumnIndex := -1
	log.Printf("Headers: %v", headers)
	for i, header := range headers {
		bytesHeader := []byte(strings.TrimSpace(header))
		bh := []byte(*c.SubmissionIDColumn)
		header = removeBOMFromString(header)
		// h := strings.TrimSpace(*c.SubmissionIDColumn)
		log.Printf("Header: %v, %v", bytesHeader, bh)
		if *c.SubmissionIDColumn == header {
			submissionIDColumnIndex = i
			break
		}
	}
	if submissionIDColumnIndex == -1 {
		(*failedImageReason)["*"] = "SubmissionID column not found : " + *c.SubmissionIDColumn
		log.Printf("SubmissionID column not found: %s", *c.SubmissionIDColumn)
		// return nil, failedImageReason
	}
	submissionIds := []string{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err == csv.ErrFieldCount {
			(*failedImageReason)["*"] = "Incorrect number of fields"
			log.Printf("Error: Incorrect number of fields")
			return nil, failedImageReason
		} else if err != nil {
			(*failedImageReason)["*"] = err.Error()
			log.Printf("Error reading CSV file: %s", err)
			return nil, failedImageReason
		}

		submissionID := record[submissionIDColumnIndex]
		if submissionID == "" {
			break
		}
		// Process the submissionID as needed
		submissionIds = append(submissionIds, submissionID)
	}
	q, close := repository.GetDBWithGen(ctx)
	// if err != nil {
	// 	(*failedImageReason)["*"] = err.Error()
	// 	log.Printf("Error getting DB connection: %s", err)
	// 	return nil, failedImageReason
	// }
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
