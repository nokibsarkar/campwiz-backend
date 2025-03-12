package distributionstrategy

import (
	"fmt"
	"log"
	"nokib/campwiz/consts"
	"nokib/campwiz/models"
	"nokib/campwiz/repository"
	idgenerator "nokib/campwiz/services/idGenerator"
	"sync"

	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/sets"
)

type RoundRobinDistributionStrategyV3 struct {
	TaskId models.IDType
}
type DistributionResultV3 struct {
	TotalWorkLoad             WorkLoadTypeV3                 `json:"totalWorkLoad"`
	AvergaeWorkLoad           WorkLoadTypeV3                 `json:"averageWorkLoad"`
	TotalWorkloadDistribution map[models.IDType]WorkLoadType `json:"totalworkloadDistribution"`
}
type WorkLoadTypeV3 int64

// Juror represents a juror with workload and ID
type JurorV3 struct {
	ID       int
	Workload WorkLoadTypeV3
}
type TaskDistributionResultV3 struct {
	TotalWorkLoad        WorkLoadTypeV3 `json:"totalWorkLoad"`
	AvergaeWorkLoad      WorkLoadTypeV3 `json:"averageWorkLoad"`
	WorkloadDistribution []WorkLoadType `json:"workloadDistribution"`
}

// MinimumWorkloadHeap is a priority queue of Jurors
type MinimumWorkloadHeapV3 []JurorV3

func (h MinimumWorkloadHeapV3) Len() int           { return len(h) }
func (h MinimumWorkloadHeapV3) Less(i, j int) bool { return h[i].Workload < h[j].Workload }
func (h MinimumWorkloadHeapV3) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinimumWorkloadHeapV3) Push(x any) {
	*h = append(*h, x.(JurorV3))
}
func (h *MinimumWorkloadHeapV3) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
func NewRoundRobinDistributionStrategyV3(taskId models.IDType) *RoundRobinDistributionStrategyV3 {
	return &RoundRobinDistributionStrategyV3{
		TaskId: taskId,
	}
}

func createMissingEvaluations(tx *gorm.DB, taskDistributionID models.IDType, evtype models.EvaluationType, round *models.Round, req []models.Submission) (int, error) {
	evaluations := []models.Evaluation{}
	for _, submission := range req {
		rest := round.Quorum - submission.AssignmentCount
		for _ = range rest {
			evID := idgenerator.GenerateID("e")
			evaluation := models.Evaluation{
				SubmissionID:       submission.SubmissionID,
				EvaluationID:       evID,
				Type:               evtype,
				DistributionTaskID: taskDistributionID,
				ParticipantID:      submission.ParticipantID,
				RoundID:            round.RoundID,
			}
			evaluations = append(evaluations, evaluation)
		}
		if res := tx.Updates(&models.Submission{
			SubmissionID:    submission.SubmissionID,
			AssignmentCount: submission.AssignmentCount + rest,
		}); res.Error != nil {
			return 0, res.Error
		}
	}
	if len(evaluations) == 0 {
		return 0, nil
	}
	res := tx.Create(&evaluations)
	if res.Error != nil {
		return 0, res.Error
	}
	return len(evaluations), nil
}

// distributeImagesBalanced distributes images among jurors while balancing workload
func distributeImagesBalancedConcurrentv3(numberOfImages, numberOfJury, distinctJuryPerImage int, alreadyWorkload []WorkLoadTypeV3) ([]sets.Set[int], error) {
	if numberOfJury < distinctJuryPerImage {
		return nil, fmt.Errorf("number of jurors must be greater than or equal to distinct jurors per image")
	}
	log.Println("Concurrently: Number of images: ", numberOfImages)
	log.Println("Number of jurors: ", numberOfJury)
	log.Println("Distinct jurors per image: ", distinctJuryPerImage)
	log.Println("Already workload: ", alreadyWorkload)
	// Min-heap to track least-loaded jurors
	hp := NewNokibPriorityQueue(&MinimumWorkloadHeapV3{})
	// Initialize the heap with jurors and their workloads
	for i := range numberOfJury {
		hp.Push(JurorV3{i, alreadyWorkload[i]})
	}
	assignments := make([]sets.Set[int], numberOfImages)
	batchCount := consts.Config.Distribution.MaximumBatchCount
	batchSize := numberOfImages / batchCount
	if batchSize < consts.Config.Distribution.MinimumBatchSize {
		batchSize = consts.Config.Distribution.MinimumBatchSize
		batchCount = numberOfImages / batchSize
	}
	if numberOfImages%batchSize != 0 {
		batchCount++
	}
	log.Printf("Batch count: %d\n", batchCount)
	var wg sync.WaitGroup
	// Mutual exclusion lock to protect the heap
	// This protects the heap for
	var perImageMutex sync.RWMutex
	wg.Add(batchCount)
	for currentBatch := range batchCount {
		start := currentBatch * batchSize
		end := min((currentBatch+1)*batchSize, numberOfImages)
		log.Println("Processing batch of images from ", start, " to ", end)
		go func() {
			defer wg.Done()
			for i := start; i < end; i++ {
				perImageMutex.Lock()
				// log.Println("Processing image: ", i)
				selectedJurySet := sets.Set[int]{} // To keep track of distinct jurors
				// Pick K distinct least-loaded jurors
				for len(selectedJurySet) < distinctJuryPerImage {
					juror := hp.Pop().(JurorV3)
					// log.Println("Juror: ", juror)
					selectedJurySet.Insert(juror.ID)
				}
				assignments[i] = selectedJurySet
				// Store the assigned jurors for this image
				for juror := range selectedJurySet {
					alreadyWorkload[juror]++
					newJuror := JurorV3{juror, alreadyWorkload[juror]}
					hp.Push(newJuror)
				}
				perImageMutex.Unlock()
			}
		}()
	}
	wg.Wait()
	hp.StopWatchHeapOps()
	return assignments, nil
}

func (strategy *RoundRobinDistributionStrategyV3) AssignJuries(tx *gorm.DB, round *models.Round, juries []models.Role) (evaluationCount int, err error) {
	submission_repo := repository.NewSubmissionRepository()
	submissions, err := submission_repo.ListAllSubmissions(tx.Where("assignment_count < ?", round.Quorum), &models.SubmissionListFilter{
		RoundID:    round.RoundID,
		CampaignID: round.CampaignID,
	})
	if err != nil {
		return 0, err
	}
	success, err := createMissingEvaluations(tx, strategy.TaskId, models.EvaluationTypeScore, round, submissions)
	if err != nil {
		return 0, err
	}
	return success, nil
}
