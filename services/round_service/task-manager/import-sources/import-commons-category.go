package importsources

import (
	"context"
	"log"
	"maps"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	"nokib/campwiz/services/round_service"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

type CommonsCategoryListSource struct {
	Categories           []string
	currentCategoryIndex int
	lastPageID           uint64
	round                *models.Round
	commons_repo         *repository.CommonsRepository
}

func (t *ImporterServer) ImportFromCommonsCategory(ctx context.Context, req *models.ImportFromCommonsCategoryRequest) (*models.ImportResponse, error) {
	log.Printf("ImportFromCommonsCategory %v", req)
	go t.importFromCommonsCategory(req.CommonsCategory, req.TaskId, req.RoundId)
	return &models.ImportResponse{
		TaskId:  req.TaskId,
		RoundId: req.RoundId,
	}, nil
}
func (t *ImporterServer) importFromCommonsCategory(categories []string, taskId string, roundId string) {
	/** Open Database Connection */
	conn, close, err := repository.GetDB()
	if err != nil {
		log.Println("Error getting DB: ", err)
		return
	}
	defer close()
	// define all the repositories
	task_repo := repository.NewTaskRepository()
	round_repo := repository.NewRoundRepository()
	// Fetch the task
	task, err := task_repo.FindByID(conn, models.IDType(taskId))
	if err != nil {
		log.Println("Error fetching task: ", err)
		return
	}
	if task == nil {
		log.Println("Task not found")
		return
	}
	// Fetch the round
	round, err := round_repo.FindByID(conn.Preload("Campaign"), models.IDType(roundId))
	if err != nil {
		log.Println("Error fetching round: ", err)
		return
	}
	if round.LatestDistributionTaskID != nil && *round.LatestDistributionTaskID != task.TaskID {
		// log.Println("Task is not the latest task for the round")
		// task.Status = models.TaskStatusFailed
		// return
	}

	commonsCategoryLister := NewCommonsCategoryListSource(categories, round)
	if commonsCategoryLister == nil {
		log.Println("Error creating commons category lister")
		task.Status = models.TaskStatusFailed
		return
	}
	successCount := 0
	failedCount := 0
	currentRoundStatus := round.Status
	{
		// Update the round status to importing
		round.LatestDistributionTaskID = &task.TaskID
		round.Status = models.RoundStatusImporting
		if res := conn.Updates(&models.Round{
			RoundID:                  round.RoundID,
			Status:                   models.RoundStatusImporting,
			LatestDistributionTaskID: &task.TaskID,
		}); res.Error != nil {
			log.Println("Error updating round status: ", res.Error)
			task.Status = models.TaskStatusFailed
			return
		}
		defer func() {
			conn.Updates(&models.Round{
				RoundID: round.RoundID,
				Status:  currentRoundStatus,
			})
		}()
	}
	log.Printf("Importing images for round %v\n", round.Campaign)
	FailedImages := &map[string]string{}
	technicalJudge := round_service.NewTechnicalJudgeService(round, round.Campaign)
	if technicalJudge == nil {
		log.Println("Error creating technical judge service")
		task.Status = models.TaskStatusFailed
		return
	}
	user_repo := repository.NewUserRepository()
	for {
		successBatch, failedBatch := commonsCategoryLister.ImportImageResults(FailedImages)
		if failedBatch != nil {
			task.FailedCount = len(*failedBatch)
			*task.FailedIds = datatypes.NewJSONType(*failedBatch)
		}
		if successBatch == nil {
			break
		}
		images := []models.MediaResult{}
		log.Println("Processing batch of images : ", len(successBatch))
		for _, image := range successBatch {
			// not allowed to submit images
			reason := technicalJudge.RejectReason(image)
			if reason != "" {
				log.Printf("File %s not allowed to submit: %s\n", image.Name, reason)
				(*FailedImages)[image.Name] = reason
				continue
			}
			images = append(images, image)
		}
		successCount += len(images)
		task.SuccessCount = successCount
		participants := map[models.WikimediaUsernameType]models.IDType{}
		for _, image := range images {
			participants[image.UploaderUsername] = idgenerator.GenerateID("u")
		}
		username2IdMap, err := user_repo.EnsureExists(conn, participants)
		if err != nil {
			log.Println("Error ensuring users exist: ", err)
			conn.Rollback()
			task.Status = models.TaskStatusFailed
			return
		}
		submissionCount := 0
		submissions := []models.Submission{}
		for _, image := range images {
			uploaderId := username2IdMap[image.UploaderUsername]
			sId := types.SubmissionIDType(idgenerator.GenerateID("s"))
			submission := models.Submission{
				SubmissionID:      sId,
				PageID:            image.PageID,
				Name:              image.Name,
				CampaignID:        *task.AssociatedCampaignID,
				URL:               image.URL,
				Author:            image.UploaderUsername,
				SubmittedByID:     uploaderId,
				ParticipantID:     uploaderId,
				SubmittedAt:       image.SubmittedAt,
				CreatedAtExternal: &image.SubmittedAt,
				RoundID:           round.RoundID,
				ImportTaskID:      task.TaskID,
				MediaSubmission: models.MediaSubmission{
					MediaType:   models.MediaType(image.MediaType),
					License:     strings.ToUpper(image.License),
					CreditHTML:  image.CreditHTML,
					Description: image.Description,
					AudioVideoSubmission: models.AudioVideoSubmission{
						Duration: image.Duration,
						Size:     image.Size,
						Bitrate:  0,
					},
					ImageSubmission: models.ImageSubmission{
						Width:      image.Width,
						Height:     image.Height,
						Resolution: image.Resolution,
					},
				},
			}
			if image.ThumbURL != nil {
				submission.ThumbURL = *image.ThumbURL
				submission.ThumbWidth = *image.ThumbWidth
				submission.ThumbHeight = *image.ThumbHeight
			}
			submissions = append(submissions, submission)
			submissionCount++
			if submissionCount%1000 == 0 {
				log.Println("Saving batch of submissions")

				res := conn.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(submissions)
				if res.Error != nil {
					conn.Rollback()
					task.Status = models.TaskStatusFailed
					log.Println("Error saving submissions: ", res.Error)
					return
				}
				submissions = []models.Submission{}
			}
		}
		if len(submissions) > 0 {
			log.Println("Saving remaining submissions")
			res := conn.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(submissions)
			if res.Error != nil {
				conn.Rollback()
				task.Status = models.TaskStatusFailed
				log.Println("Error saving submissions: ", res.Error)
				return
			}
		}
		*task.FailedIds = datatypes.NewJSONType(*failedBatch)
		res := conn.Save(task)
		if res.Error != nil {
			log.Println("Error saving task: ", res.Error)
			task.Status = models.TaskStatusFailed
			conn.Rollback()
			return
		}
	}

	// commonsRepo := repository.NewCommonsRepository()
	// submissionRepo := repository.NewSubmissionRepository()
	// pageids, err := submissionRepo.GetPageIDsForWithout(conn, *task.AssociatedRoundID)
	// if err != nil {
	// 	log.Println("Error fetching page ids: ", err)
	// 	task.Status = models.TaskStatusFailed
	// 	return
	// }
	// images := commonsRepo.GetImagesThumbsFromIPageIDs(pageids)

	// tx := conn.Begin()
	// if len(images) > 0 {
	// 	for _, image := range images {
	// 		res := tx.Model(&models.Submission{}).Where(&models.Submission{PageID: image.PageID}).Updates(image)

	// 		if res.Error != nil {
	// 			log.Println("Error updating image: ", res.Error)
	// 			tx.Rollback()
	// 			task.Status = models.TaskStatusFailed
	// 			return
	// 		}
	// 	}
	// }
	// tx.Commit()
	go t.importDescriptions(round)
	{
		task.Status = models.TaskStatusSuccess
		round.LatestDistributionTaskID = nil // Reset the latest task id
	}
	if err := t.updateStatistics(conn, round, successCount, failedCount); err != nil {
		log.Println("Error updating statistics: ", err)
		task.Status = models.TaskStatusFailed
	}
}
func (t *ImporterServer) importDescriptions(round *models.Round) {
	conn, close, err := repository.GetDB()
	if err != nil {
		log.Println("Error getting DB: ", err)
		return
	}
	defer close()
	submission_repo := repository.NewSubmissionRepository()
	commons_repo := repository.NewCommonsRepository()
	lastPageID := uint64(0)
	batchSize := 1000
	lastCount := batchSize
	for lastCount == batchSize {
		log.Println("Fetching page ids without description")
		nonDescriptionPageIds, err := submission_repo.GetPageIDWithoutDescriptionByRoundID(conn, round.RoundID, lastPageID, batchSize)
		if err != nil {
			log.Println("Error fetching page ids: ", err)
			return
		}
		lastCount = len(nonDescriptionPageIds)
		if len(nonDescriptionPageIds) > 0 {
			log.Println("Updating descriptions for images")
			m := commons_repo.GetImagesDescriptionFromIPageIDs(nonDescriptionPageIds)
			for _, image := range m {
				res := conn.Where(&models.Submission{PageID: image.PageID}).Updates(models.Submission{
					MediaSubmission: models.MediaSubmission{
						Description: image.Description,
						License:     strings.ToUpper(image.License),
						CreditHTML:  image.CreditHTML,
					},
				})
				if res.Error != nil {
					log.Println("Error updating image: ", res.Error)
					return
				}
				lastPageID = image.PageID
			}
			log.Println("Updating descriptions for images done")
		}
	}
}

// ImportImageResults imports images from commons categories
// For Each invocation it will import images from a single category
// If all categories are imported it will return nil
// If there are no images in the category it will return nil
// If there are images in the category it will return the images
// If there are failed images in the category it will return the reason as value of the map
func (c *CommonsCategoryListSource) ImportImageResults(failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string) {
	if c.currentCategoryIndex < len(c.Categories) {
		category := c.Categories[c.currentCategoryIndex]
		campaign := c.round.Campaign
		successMedia, currentfailedImages, lastPageID := c.commons_repo.GetImagesFromCommonsCategories2(category, c.lastPageID, c.round, campaign.StartDate, campaign.EndDate)
		if lastPageID == 0 {
			c.currentCategoryIndex++
		}
		c.lastPageID = lastPageID
		maps.Copy(*failedImageReason, currentfailedImages)
		return successMedia, failedImageReason
	}

	return nil, failedImageReason
}

func NewCommonsCategoryListSource(categories []string, round *models.Round) *CommonsCategoryListSource {
	ct := []string{}
	for _, category := range categories {
		kt := strings.Replace(category, " ", "_", -1)
		kt = strings.Replace(kt, "Category:", "", -1)
		ct = append(ct, kt)
	}
	return &CommonsCategoryListSource{
		Categories:           ct,
		currentCategoryIndex: 0,
		lastPageID:           0,
		commons_repo:         repository.NewCommonsRepository(),
		round:                round,
	}
}
