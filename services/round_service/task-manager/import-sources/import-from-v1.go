package importsources

import (
	"context"
	"fmt"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type V1Source struct {
	fileName            string
	fromCampaignId      int
	toRoundId           string
	conn                *gorm.DB
	done                bool
	lastSubmissionId    int
	submissionId2pageId map[int]int
	participants        map[models.WikimediaUsernameType]struct{}
}

func NewV1Source(fileName string, fromCampaignId int, toRoundId string) *V1Source {
	return &V1Source{
		fileName:            fileName,
		fromCampaignId:      fromCampaignId,
		toRoundId:           toRoundId,
		lastSubmissionId:    0,
		submissionId2pageId: make(map[int]int),
		done:                false,
		conn:                nil,
		participants:        make(map[models.WikimediaUsernameType]struct{}),
	}
}
func (t *ImporterServer) ImportFromCampWizV1(ctx context.Context, req *models.ImportFromCampWizV1Request) (*models.ImportResponse, error) {
	source := NewV1Source(
		req.Path,
		int(req.CampaignId),
		req.RoundId,
	)
	go t.importFrom(context.Background(), source, req.TaskId, req.RoundId)
	return &models.ImportResponse{}, nil

}

func (s *V1Source) getSQLiteConn() (*gorm.DB, error) {
	dsn := s.fileName
	conn, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

const batchSize = 10000

func (t *V1Source) ImportImageResults(ctx context.Context, currentRound *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	if t.done {
		return nil, failedImageReason
	}
	if t.conn == nil {
		conn, err := t.getSQLiteConn()
		if err != nil {
			if failedImageReason != nil {
				*failedImageReason = map[string]string{"error": "Failed to connect to SQLite database: " + err.Error()}
			}
			return nil, failedImageReason
		}
		t.conn = conn
	}
	startIndex := t.lastSubmissionId
	raw_data := []models.CampWizV1Submission{}
	results := []models.MediaResult{}
	res := t.conn.Table("submission").
		Where("campaign_id = ?", t.fromCampaignId).
		Where("submission_id > ?", startIndex).
		Order("id").Limit(batchSize).Scan(&raw_data)
	if res.Error != nil && res.Error.Error() == "record not found" {
		// No more submissions to process
		t.done = true
		return nil, failedImageReason
	}
	if res.Error != nil && res.Error.Error() != "record not found" {
		if failedImageReason != nil {
			*failedImageReason = map[string]string{"error": "Failed to fetch submissions: " + res.Error.Error()}
		}
		return nil, failedImageReason
	}
	t.lastSubmissionId += len(raw_data)
	if res.Error != nil {
		if failedImageReason != nil {
			*failedImageReason = map[string]string{"error": "Failed to fetch submissions: " + res.Error.Error()}
		}
		return nil, failedImageReason
	}
	for _, submission := range raw_data {
		result := models.MediaResult{
			PageID:           uint64(submission.PageID),
			Name:             submission.Title,
			SubmittedAt:      submission.SubmittedAt,
			UploaderUsername: models.WikimediaUsernameType(submission.CreatedByUsername),
			Size:             uint64(submission.TotalBytes),
			MediaType:        string(models.MediaTypeArticle),
		}
		t.submissionId2pageId[submission.SubmissionID] = submission.PageID
		t.participants[models.WikimediaUsernameType(submission.CreatedByUsername)] = struct{}{}
		results = append(results, result)
	}
	t.done = len(results) < batchSize
	return results, failedImageReason
}

func (t *V1Source) PostProcess(ctx context.Context, tx *gorm.DB, currentRound *models.Round, task *models.Task, importMap map[uint64]types.SubmissionIDType, newlyCreatedUsers map[models.WikimediaUsernameType]models.IDType) error {
	conn := t.conn
	lastEvaluationIndex := 0
	type Jury struct {
		UserID int                          `gorm:"column:user_id"`
		Name   models.WikimediaUsernameType `gorm:"column:username"`
	}
	jury := []Jury{}
	res := conn.Table("jury").Select(("user_id, username")).
		Where("campaign_id = ?", t.fromCampaignId).
		Where("allowed = ?", true).
		Order("user_id").Scan(&jury)
	if res.Error != nil {
		return fmt.Errorf("failed to fetch jury members: %w", res.Error)
	}
	if len(jury) == 0 {
		return fmt.Errorf("no jury members found for campaign ID %d", t.fromCampaignId)
	}
	juryUsernames := []models.WikimediaUsernameType{}
	juryId2NameMap := map[int]models.WikimediaUsernameType{}
	for _, j := range jury {
		juryUsernames = append(juryUsernames, j.Name)
		juryId2NameMap[j.UserID] = j.Name
	}

	role_repo := repository.NewRoleRepository()
	juryRoleType := models.RoleTypeJury
	juryRoles, err := role_repo.FindRolesByUsername(tx.Preload("User"), juryUsernames, &models.RoleFilter{
		RoundID: &currentRound.RoundID,
		Type:    &juryRoleType,
	})
	if err != nil {
		return fmt.Errorf("failed to find jury roles: %w", err)
	}
	targetUsernames := append([]models.WikimediaUsernameType{}, juryUsernames...)
	for username := range t.participants {
		targetUsernames = append(targetUsernames, username)
	}
	user_repo := repository.NewUserRepository()
	nameMap, err := user_repo.FetchExistingUsernames(tx, targetUsernames)
	if err != nil {
		return fmt.Errorf("failed to fetch existing usernames: %w", err)
	}
	roleMap := map[models.IDType]models.IDType{}
	for _, role := range juryRoles {
		roleMap[role.UserID] = role.RoleID
	}
	// Get Submission ID to Page ID mapping
	for {
		evaluations := []models.Evaluation{}
		previousEvaluations := []models.CampWizV1Evaluation{}
		res := conn.Table("jury_vote").
			Where("campaign_id = ?", t.fromCampaignId).
			Where("id > ?", lastEvaluationIndex).
			Order("id").Limit(batchSize).Scan(&previousEvaluations)
		if res.Error != nil {
			return res.Error
		}
		if len(previousEvaluations) == 0 {
			break
		}
		for _, eval := range previousEvaluations {
			pageId, exists := t.submissionId2pageId[eval.SubmissionID]
			if !exists {
				continue // Skip if the submission ID does not exist in the map
			}
			ourSubmissionId, ok := importMap[uint64(pageId)]
			if !ok {
				continue // Skip if the page ID is not in the import map
			}
			juryUsername, ok := juryId2NameMap[eval.JuryID]
			if !ok {
				continue // Skip if the jury ID does not exist in the map
			}
			score := models.ScoreType(float64(eval.Vote) * 100.0)
			// asume the jury username is in the nameMap
			juryUserId := nameMap[juryUsername]
			juryRoleId := roleMap[juryUserId]
			evaluation := models.Evaluation{
				EvaluationID: idgenerator.GenerateID("e"),
				SubmissionID: ourSubmissionId,
				JudgeID:      &juryRoleId,
				RoundID:      currentRound.RoundID,
				Score:        &score,
				AssignedAt:   &eval.CreatedAt,
				EvaluatedAt:  &eval.CreatedAt,
			}
			evaluations = append(evaluations, evaluation)
		}
		res = tx.Create(&evaluations)
		if res.Error != nil {
			return res.Error
		}
		lastEvaluationIndex = previousEvaluations[len(previousEvaluations)-1].EvaluationId
	}
	return nil
}
