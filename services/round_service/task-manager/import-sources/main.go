package importsources

import (
	"context"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	"nokib/campwiz/services/round_service"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ImporterServer struct {
	models.UnimplementedImporterServer
}

// ImportService is an interface for importing data from different sources
// All the importer services should implement this interface
type IImportSource interface {
	// This method would be called in a loop to fetch each batch of images
	// It should return the images that were successfully imported and the images that failed to import
	// If there are no images to import it should return nil
	// If there are failed images it should return the reason as a map
	ImportImageResults(ctx context.Context, round *models.Round, failedImageReason *map[string]string) ([]models.MediaResult, *map[string]string)
}
type IDistributionStrategy interface {
	AssignJuries(tx *gorm.DB, round *models.Round, juries []models.Role) (success int, fail int, err error)
}

func (imp *ImporterServer) updateStatistics(tx *gorm.DB, round *models.Round, successCount, failedCount int) error {
	type Result struct {
		TotalSubmissions          int
		TotalEvaluatedSubmissions int
	}
	var result Result
	q := query.Use(tx)
	Submission := q.Submission
	err := Submission.Select(Submission.SubmissionID.Count().As("TotalSubmissions"), Submission.AssignmentCount.Sum().
		As("TotalEvaluatedSubmissions")).Where(Submission.RoundID.Eq(round.RoundID.String())).Scan(&result)
	if err != nil {
		return err
	}
	res := tx.Updates(&models.Round{
		RoundID:                   round.RoundID,
		TotalSubmissions:          result.TotalSubmissions,
		TotalEvaluatedSubmissions: result.TotalEvaluatedSubmissions,
	})
	return res.Error
}
func NewImporterServer() *ImporterServer {
	return &ImporterServer{}
}
func (t *ImporterServer) importFrom(ctx context.Context, source IImportSource, taskId string, currentRoundId string) {

	log.Printf("Importing from source for task %s in round %s\n", taskId, currentRoundId)
	/** Open Database Connection */
	conn, close, err := repository.GetDB(ctx)
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
	log.Printf("Task fetched: %s %v\n", taskId, task)
	if task == nil {
		log.Println("Task not found")
		return
	}

	log.Printf("Task found: %v\n", task)
	// Fetch the currentRound
	currentRound, err := round_repo.FindByID(conn.Preload("Campaign"), models.IDType(currentRoundId))
	if err != nil {
		log.Println("Error fetching round: ", err)
		return
	}
	log.Printf("Current round: %v\n", currentRound)
	// if round.LatestDistributionTaskID != nil && *round.LatestDistributionTaskID != task.TaskID {
	// 	log.Println("Task is not the latest task for the round")
	// 	task.Status = models.TaskStatusFailed
	// 	return
	// }

	if source == nil {
		log.Println("Error creating commons category lister")
		task.Status = models.TaskStatusFailed
		return
	}
	successCount := 0
	failedCount := 0
	currentRoundStatus := currentRound.Status
	{
		log.Printf("Starting import for round %s with task %s\n", currentRound.RoundID, task.TaskID)
		// Update the round status to importing
		currentRound.LatestDistributionTaskID = &task.TaskID
		currentRound.Status = models.RoundStatusImporting
		if res := conn.Updates(&models.Round{
			RoundID:                  currentRound.RoundID,
			Status:                   models.RoundStatusImporting,
			LatestDistributionTaskID: &task.TaskID,
		}); res.Error != nil {
			log.Println("Error updating round status: ", res.Error)
			task.Status = models.TaskStatusFailed
			return
		}
		defer func() {
			log.Println("Finalizing task and round status")
			res := conn.Updates(&models.Round{
				RoundID: currentRound.RoundID,
				Status:  currentRoundStatus,
			})
			if res.Error != nil {
				log.Println("Error updating round status: ", res.Error)
				task.Status = models.TaskStatusFailed
				return
			}
			res = conn.Updates(&models.Task{
				TaskID:         task.TaskID,
				Status:         task.Status,
				SuccessCount:   successCount,
				FailedCount:    failedCount,
				RemainingCount: 0,
			})
			if res.Error != nil {
				log.Println("Error updating task status: ", res.Error)
				task.Status = models.TaskStatusFailed
				return
			}
		}()
	}
	FailedImages := &map[string]string{}
	technicalJudge := round_service.NewTechnicalJudgeService(currentRound, currentRound.Campaign)
	if technicalJudge == nil {
		log.Println("Error creating technical judge service")
		task.Status = models.TaskStatusFailed
		return
	}
	user_repo := repository.NewUserRepository()
	for {
		log.Println("Importing images from source")
		successBatch, failedBatch := source.ImportImageResults(ctx, currentRound, FailedImages)
		log.Printf("Received batch of images: %d success, %d failed\n", len(successBatch), len(*FailedImages))
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
				RoundID:           currentRound.RoundID,
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
	go t.importDescriptions(ctx, currentRound)
	{
		task.Status = models.TaskStatusSuccess
		currentRound.LatestDistributionTaskID = nil // Reset the latest task id
	}
	if err := t.updateStatistics(conn, currentRound, successCount, failedCount); err != nil {
		log.Println("Error updating statistics: ", err)
		task.Status = models.TaskStatusFailed
	}
	log.Printf("Round %s imported successfully\n", currentRound.RoundID)
}
func (t *ImporterServer) importDescriptions(ctx context.Context, round *models.Round) {
	conn, close, err := repository.GetDB(ctx)
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
