package models

import (
	"gorm.io/gorm"
)

type RoundStatisticsView struct {
	RoundName        string    `json:"roundName"`
	RoundID          IDType    `json:"roundId"`
	ParticipantID    IDType    `json:"participantId"`
	Username         string    `json:"participantName"`
	TotalSubmissions int       `json:"totalSubmissions"`
	TotalScore       ScoreType `json:"totalScore"`
}

func (r RoundStatisticsView) TableName() string {
	return "round_statistics_view"
}
func (r RoundStatisticsView) GetQuery(db *gorm.DB) *gorm.DB {
	// This view is used to get the statistics of each round for each participant
	return db.Select(
		"users.username AS participant_name",
		"rounds.name AS round_name",
		"rounds.round_id AS round_id",
		"submissions.participant_id AS participant_id",
		"COUNT(submissions.submission_id) AS total_submissions",
		"SUM(submissions.score) AS total_score",
	).Table("submissions").
		Joins("LEFT JOIN rounds ON submissions.round_id = rounds.round_id").
		Joins("LEFT JOIN users ON submissions.participant_id = users.user_id").
		Group("submissions.participant_id").
		Order("rounds.round_id")

}

type RoundResult struct {
	AverageScore    float64 `json:"averageScore"`
	SubmissionCount int     `json:"submissionCount"`
}
type RoundStatistics struct {
	RoundID         IDType
	AssignmentCount int
	EvaluationCount int
}
type RoundStatisticsFetcher interface {
	// SELECT SUM(`assignment_count`) AS `AssignmentCount`, SUM(`evaluation_count`) AS EvaluationCount, `round_id` AS `round_id` FROM `submissions` WHERE `round_id` = @round_id
	FetchByRoundID(round_id string) ([]RoundStatistics, error)
	// UPDATE rounds,
	// (SELECT s.round_id, COUNT(*) AS TotalSubmissions, SUM(s.assignment_count) AS TotalAssignments,
	// SUM(s.evaluation_count) AS TotalEvaluatedAssignments, SUM(CASE WHEN s.evaluation_count >= r.quorum THEN 1 ELSE 0 END)
	// AS TotalEvaluatedSubmissions, SUM(s.score) AS TotalScore FROM submissions s FORCE INDEX (idx_submissions_round_id)
	// JOIN rounds r ON s.round_id = r.round_id WHERE s.round_id = @round_id LIMIT 1) AS s_data
	// SET rounds.total_submissions = s_data.TotalSubmissions,
	// rounds.total_assignments = s_data.TotalAssignments, rounds.total_evaluated_assignments = s_data.TotalEvaluatedAssignments,
	//  rounds.total_evaluated_submissions = s_data.TotalEvaluatedSubmissions, rounds.total_score = s_data.TotalScore
	// WHERE rounds.round_id = @round_id;
	UpdateByRoundID(round_id string) error
	// SELECT
	// users.username AS participant_name,
	// rounds.name AS round_name,
	// rounds.round_id AS round_id,
	// submissions.participant_id AS participant_id,
	// COUNT(submissions.submission_id) AS total_submissions,
	// SUM(submissions.score) AS total_score
	// FROM `submissions` FORCE INDEX (idx_submissions_round_id)
	// LEFT JOIN rounds ON submissions.round_id = rounds.round_id
	// LEFT JOIN users ON submissions.participant_id = users.user_id
	// WHERE submissions.round_id IN (@round_ids)
	// GROUP BY `submissions`.`participant_id`
	// ORDER BY total_score DESC, total_submissions DESC;
	FetchUserStatisticsByRoundIDs(round_ids []string) ([]RoundStatisticsView, error)
}
