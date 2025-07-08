package importsources

import (
	"context"
	"errors"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	"nokib/campwiz/services/round_service"
	"strings"

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
type IImportSourceWithPostProcessing interface {
	IImportSource
	// This method is called after the import is done to perform any post-processing
	PostProcess(ctx context.Context, conn *gorm.DB, round *models.Round, task *models.Task, pageIdMap map[uint64]types.SubmissionIDType, newlyCreatedUsers map[models.WikimediaUsernameType]models.IDType) error
}

type IDistributionStrategy interface {
	AssignJuries(tx *gorm.DB, round *models.Round, juries []models.Role) (success int, fail int, err error)
}

func (imp *ImporterServer) updateRoundStatistics(tx *gorm.DB, round *models.Round, successCount, failedCount int) error {
	q := query.Use(tx)
	if err := q.SubmissionStatistics.TriggerByRoundId(round.RoundID.String()); err != nil {
		log.Println("Error triggering submission statistics: ", err)
		return err
	}
	return q.RoundStatistics.UpdateByRoundID(round.RoundID.String())

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
		err = errors.New("task.notFound")
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
		err = errors.New("source.notFound")
		log.Println("Error creating commons category lister")
		task.Status = models.TaskStatusFailed
		return
	}
	successCount := 0
	failedCount := 0
	currentRoundStatus := currentRound.Status
	log.Printf("Starting import for round %s with task %s\n", currentRound.RoundID, task.TaskID)
	// Update the round status to importing
	currentRound.LatestDistributionTaskID = &task.TaskID
	currentRound.Status = models.RoundStatusImporting
	if res := conn.Updates(&models.Round{
		RoundID:                  currentRound.RoundID,
		Status:                   models.RoundStatusImporting,
		LatestDistributionTaskID: &task.TaskID,
	}); res.Error != nil {
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
			err = res.Error
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
			err = res.Error
			log.Println("Error updating task status: ", res.Error)
			task.Status = models.TaskStatusFailed
			return
		}
	}()
	tx := conn.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction: ", tx.Error)
		return
	}
	defer func() {
		if err != nil {
			log.Println("Rolling back transaction due to error: ", err)
			if r := tx.Rollback(); r.Error != nil {
				log.Println("Error rolling back transaction: ", r.Error)
			}
		} else {
			log.Println("Committing transaction")
			if r := tx.Commit(); r.Error != nil {
				log.Println("Error committing transaction: ", r.Error)
			}
		}
	}()
	FailedImages := &map[string]string{}
	technicalJudge := round_service.NewTechnicalJudgeService(currentRound, currentRound.Campaign)
	if technicalJudge == nil {
		err = errors.New("technicalJudgeServicenotFound")
		log.Println("Error creating technical judge service")
		task.Status = models.TaskStatusFailed
		return
	}
	user_repo := repository.NewUserRepository()
	pageIdMap := map[uint64]types.SubmissionIDType{}
	newlyCreatedUsers := map[models.WikimediaUsernameType]models.IDType{}
	var username2IdMap map[models.WikimediaUsernameType]models.IDType
	for {
		log.Println("Importing images from source")
		successBatch, failedBatch := source.ImportImageResults(ctx, currentRound, FailedImages)
		log.Printf("Received batch of images: %d success, %d failed\n", len(successBatch), len(*FailedImages))
		if failedBatch != nil {
			task.FailedCount = len(*failedBatch)
			// *task.FailedIds = datatypes.NewJSONType(*failedBatch)
		}
		if len(successBatch) == 0 {
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
			if image.CreatedByUsername != "" {
				participants[image.CreatedByUsername] = idgenerator.GenerateID("u")
			}
			if image.SubmittedByUsername != "" {
				participants[image.SubmittedByUsername] = idgenerator.GenerateID("u")
			}
		}
		username2IdMap, err = user_repo.EnsureExists(tx, participants)
		if err != nil {
			log.Println("Error ensuring users exist: ", err)
			task.Status = models.TaskStatusFailed
			return
		}
		submissionCount := 0
		submissions := []models.Submission{}
		for _, image := range images {
			submitterId := username2IdMap[image.SubmittedByUsername]
			creatorId := username2IdMap[image.CreatedByUsername]
			newlyCreatedUsers[image.SubmittedByUsername] = submitterId
			newlyCreatedUsers[image.CreatedByUsername] = creatorId
			sId := types.SubmissionIDType(idgenerator.GenerateID("s"))
			submission := models.Submission{
				SubmissionID:      sId,
				PageID:            image.PageID,
				Name:              image.Name,
				CampaignID:        *task.AssociatedCampaignID,
				URL:               image.URL,
				Author:            image.CreatedByUsername,
				SubmittedByID:     submitterId,
				ParticipantID:     submitterId,
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
			pageIdMap[image.PageID] = sId
			submissions = append(submissions, submission)
			submissionCount++
			if submissionCount%1000 == 0 {
				log.Println("Saving batch of submissions")
				res := tx.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(submissions)
				if res.Error != nil {
					err = res.Error
					task.Status = models.TaskStatusFailed
					log.Println("Error saving submissions: ", res.Error)
					return
				}
				submissions = []models.Submission{}
			}
		}
		if len(submissions) > 0 {
			submissionIds := make([]string, 0, len(pageIdMap))
			for _, submissionId := range submissions {
				submissionIds = append(submissionIds, string(submissionId.SubmissionID))
			}
			log.Printf("Total submissions imported: %d, failed: %d\n", successCount, len(*FailedImages))
			q := query.Use(tx)
			RES, err := q.Submission.Where(q.Submission.SubmissionID.In(submissionIds...)).Find()
			if err != nil {
				log.Println("Error fetching submissions: ", err)
				task.Status = models.TaskStatusFailed
				return
			}
			log.Printf("Fetched %d submissions from the database\n", len(RES))
			log.Println("Saving remaining submissions")
			res := tx.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(submissions)
			if res.Error != nil {
				task.Status = models.TaskStatusFailed
				log.Println("Error saving submissions: ", res.Error)
				return
			}
		}

		// *task.FailedIds = datatypes.NewJSONType(*failedBatch)
		res := tx.Save(task)
		if res.Error != nil {
			err = res.Error
			log.Println("Error saving task: ", res.Error)
			task.Status = models.TaskStatusFailed
			return
		}

	}

	if p, ok := source.(IImportSourceWithPostProcessing); ok {
		log.Println("Post-processing import")
		err = p.PostProcess(ctx, tx, currentRound, task, pageIdMap, newlyCreatedUsers)
		if err != nil {
			log.Println("Error in post-processing import: ", err)
			task.Status = models.TaskStatusFailed
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
	// go t.importDescriptions(ctx, currentRound)
	task.Status = models.TaskStatusSuccess
	currentRound.LatestDistributionTaskID = nil // Reset the latest task id
	err = t.updateRoundStatistics(tx, currentRound, successCount, failedCount)
	if err != nil {
		log.Println("Error updating statistics: ", err)
		task.Status = models.TaskStatusFailed
	}
	log.Printf("Round %s imported successfully\n", currentRound.RoundID)
	defer t.importDescriptions(ctx, currentRound)
	err = nil
}
func (t *ImporterServer) importDescriptions(ctx context.Context, round *models.Round) {
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		log.Println("Error getting DB: ", err)
		return
	}
	defer close()
	submission_repo := repository.NewSubmissionRepository()
	commons_repo := repository.NewCommonsRepository(nil)
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
