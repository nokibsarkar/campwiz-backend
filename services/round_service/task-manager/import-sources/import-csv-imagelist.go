package importsources

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"nokib/campwiz/models"
	"os"
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
	go t.importFrom(source, req.TaskId, req.RoundId)
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
func (c *CSVListSource) ImportImageResults(currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
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
		return c.importUsingSubmissionId(reader, failedImageReason)
	}
	return nil, failedImageReason
}
func (c *CSVListSource) importUsingSubmissionId(reader *csv.Reader, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	headers, err := reader.Read()
	if err != nil {
		(*failedImageReason)["*"] = err.Error()
		log.Printf("Error reading CSV file: %s", err)
		return nil, failedImageReason
	}
	submissionIDColumnIndex := -1
	for i, header := range headers {
		if *c.SubmissionIDColumn == header {
			submissionIDColumnIndex = i
			break
		}
	}
	if submissionIDColumnIndex == -1 {
		(*failedImageReason)["*"] = "SubmissionID column not found : " + *c.SubmissionIDColumn
		log.Printf("SubmissionID column not found: %s", *c.SubmissionIDColumn)
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
	log.Printf("Submission IDs: %v", submissionIds)
	return nil, failedImageReason
}
