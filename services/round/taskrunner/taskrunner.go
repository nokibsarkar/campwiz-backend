package taskrunner

import (
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/models/types"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	rnd "nokib/campwiz/services/round"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ImportService is an interface for importing data from different sources
// All the importer services should implement this interface
type IImportSource interface {
	// This method would be called in a loop to fetch each batch of images
	// It should return the images that were successfully imported and the images that failed to import
	// If there are no images to import it should return nil
	// If there are failed images it should return the reason as a map
	ImportImageResults(failedImageReason *map[string]string) ([]models.ImageResult, *map[string]string)
}
type IDistributionStrategy interface {
	AssignJuries(tx *gorm.DB, round *models.Round, juries []models.Role) (success int, fail int, err error)
}

type TaskRunner struct {
	TaskId               models.IDType
	ImportSource         IImportSource
	DistributionStrategy IDistributionStrategy
}

func NewImportTaskRunner(taskId models.IDType, importService IImportSource) *TaskRunner {
	return &TaskRunner{
		TaskId:       taskId,
		ImportSource: importService,
	}
}
func NewDistributionTaskRunner(taskId models.IDType, strategy IDistributionStrategy) *TaskRunner {
	return &TaskRunner{
		TaskId:               taskId,
		DistributionStrategy: strategy,
	}
}
func (b *TaskRunner) importImages(conn *gorm.DB, task *models.Task) (successCount, failedCount int) {
	round_repo := repository.NewRoundRepository()
	round, err := round_repo.FindByID(conn, *task.AssociatedRoundID)
	if err != nil {
		log.Println("Error fetching round: ", err)
		return
	}
	if round.LatestDistributionTaskID != nil && *round.LatestDistributionTaskID != task.TaskID {
		// log.Println("Task is not the latest task for the round")
		// task.Status = models.TaskStatusFailed
		// return
	}
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
	FailedImages := &map[string]string{}
	technicalJudge := rnd.NewTechnicalJudgeService(round)
	user_repo := repository.NewUserRepository()
	for {
		successBatch, failedBatch := b.ImportSource.ImportImageResults(FailedImages)
		if failedBatch != nil {
			task.FailedCount = len(*failedBatch)
			*task.FailedIds = datatypes.NewJSONType(*failedBatch)
		}
		if successBatch == nil {
			break
		}
		images := []models.ImageResult{}
		log.Println("Processing batch of images")
		for _, image := range successBatch {
			if technicalJudge.PreventionReason(image) != "" {
				images = append(images, image)
			}
		}
		successCount += len(images)
		task.SuccessCount = successCount
		participants := map[models.WikimediaUsernameType]models.IDType{}
		for _, image := range images {
			participants[image.UploaderUsername] = idgenerator.GenerateID("u")
		}
		perBatch := conn.Begin()
		username2IdMap, err := user_repo.EnsureExists(perBatch, participants)
		if err != nil {
			log.Println("Error ensuring users exist: ", err)
			perBatch.Rollback()
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
					ThumbURL:    image.URL,
					ThumbWidth:  image.Width,
					ThumbHeight: image.Height,
					License:     strings.ToUpper(image.License),
					CreditHTML:  image.CreditHTML,
					Description: image.Description,
					AudioVideoSubmission: models.AudioVideoSubmission{
						Duration: image.Duration,
						Size:     image.Size,
						Bitrate:  0,
					},
					ImageSubmission: models.ImageSubmission{
						Width:  image.Width,
						Height: image.Height,
					},
				},
			}
			submissions = append(submissions, submission)
			submissionCount++
			if submissionCount%200 == 0 {
				log.Println("Saving batch of submissions")

				res := perBatch.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(submissions)
				if res.Error != nil {
					perBatch.Rollback()
					task.Status = models.TaskStatusFailed
					log.Println("Error saving submissions: ", res.Error)
					return
				}
				submissions = []models.Submission{}
			}
		}
		if len(submissions) > 0 {
			log.Println("Saving remaining submissions")
			res := perBatch.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(submissions)
			if res.Error != nil {
				perBatch.Rollback()
				task.Status = models.TaskStatusFailed
				log.Println("Error saving submissions: ", res.Error)
				return
			}
		}
		*task.FailedIds = datatypes.NewJSONType(*failedBatch)
		res := perBatch.Save(task)
		if res.Error != nil {
			log.Println("Error saving task: ", res.Error)
			task.Status = models.TaskStatusFailed
			perBatch.Rollback()
			return
		}
		perBatch.Commit()
	}
	{
		task.Status = models.TaskStatusSuccess
		round.LatestDistributionTaskID = nil // Reset the latest task id
	}
	if err := b.updateStatistics(conn, round, successCount, failedCount); err != nil {
		log.Println("Error updating statistics: ", err)
		task.Status = models.TaskStatusFailed
	}
	return
}
func (b *TaskRunner) updateStatistics(tx *gorm.DB, round *models.Round, successCount, failedCount int) error {
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

func (b *TaskRunner) distributeEvaluations(tx *gorm.DB, task *models.Task) (successCount, failedCount int, err error) {
	round_repo := repository.NewRoundRepository()
	round, err := round_repo.FindByID(tx, *task.AssociatedRoundID)
	if err != nil {
		log.Println("Error fetching round: ", err)
		return
	}
	jury_repo := repository.NewRoleRepository()
	filter := &models.RoleFilter{
		RoundID:    &round.RoundID,
		CampaignID: &round.CampaignID,
	}
	j := models.RoleTypeJury
	filter.Type = &j
	// cacheDB, closeCache := cache.GetCacheDB()
	// defer closeCache()
	// _, err = cache.ExportToCache(tx, cacheDB, &models.EvaluationFilter{
	// 	// RoundID:    round.RoundID,
	// 	// CampaignID: round.CampaignID,
	// 	CommonFilter: repository.CommonFilter{
	// 		Limit: 50,
	// 	},
	// }) //, task.TaskID)
	// if err != nil {
	// 	log.Println("Error exporting to cache: ", err)
	// 	cacheDB.Rollback()
	// 	return
	// }
	// cacheDB.Commit()
	juries, err := jury_repo.ListAllRoles(tx, filter)
	if err != nil {
		log.Println("Error fetching juries: ", err)
		return
	}
	if len(juries) == 0 {
		log.Println("No juries found")
		return
	}
	log.Printf("Found %d juries\n", len(juries))

	successCount, failedCount, err = b.DistributionStrategy.AssignJuries(tx, round, juries)
	if err != nil {
		log.Println("Error assigning juries: ", err)
		tx.Rollback()
		return
	}
	return
}

func (b *TaskRunner) Run() {
	task_repo := repository.NewTaskRepository()
	conn, close := repository.GetDB()
	defer close()

	task, err := task_repo.FindByID(conn, b.TaskId)
	if err != nil {
		log.Println("Error fetching task: ", err)
		return
	}
	defer func() {
		res := conn.Updates(&task)
		if res.Error != nil {
			log.Println("Error saving task: ", res.Error)
		}
	}()
	successCount, failedCount := 0, 0
	if task.Type == models.TaskTypeImportFromCommons || task.Type == models.TaskTypeImportFromPreviousRound {
		if b.ImportSource == nil {
			log.Printf("No import source found for task %s\n", b.TaskId)
			task.Status = models.TaskStatusFailed
			return
		}
		successCount, failedCount = b.importImages(conn, task)
		task.Status = models.TaskStatusSuccess
	} else if task.Type == models.TaskTypeDistributeEvaluations {
		tx := conn.Begin()
		successCount, failedCount, err = b.distributeEvaluations(tx, task)
		if err != nil {
			log.Println("Error distributing evaluations: ", err)
			tx.Rollback()
			task.Status = models.TaskStatusFailed
			return
		}
		task.Status = models.TaskStatusSuccess
		tx.Commit()
	} else {
		log.Printf("Unknown task type %s\n", task.Type)
		task.Status = models.TaskStatusFailed
		return
	}
	task.Status = models.TaskStatusSuccess
	log.Printf("Task %s completed with %d success and %d failed\n", b.TaskId, successCount, failedCount)
}
