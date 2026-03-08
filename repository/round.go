package repository

import (
	"nokib/campwiz/models"
	"nokib/campwiz/query"

	"gorm.io/gorm"
)

type RoundRepository struct{}

func NewRoundRepository() *RoundRepository {
	return &RoundRepository{}
}
func (r *RoundRepository) Create(conn *gorm.DB, round *models.Round) (*models.Round, error) {
	result := conn.Create(round)
	if result.Error != nil {
		return nil, result.Error
	}
	return round, nil
}
func (r *RoundRepository) Update(conn *gorm.DB, round *models.Round) (*models.Round, error) {
	booleanColumns := []string{
		"is_public_jury", "is_open", "allow_jury_to_participate",
		"allow_multiple_judgement", "secret_ballot",
		"article_allow_expansions", "article_allow_creations",
	}
	if err := conn.Omit(booleanColumns...).Updates(round).Error; err != nil {
		return nil, err
	}
	// GORM's Updates skips zero-value fields; for booleans false is the zero
	// value, so we must force-update every boolean column explicitly.
	booleans := map[string]interface{}{
		"is_public_jury":            round.IsPublicJury,
		"is_open":                   round.IsOpen,
		"allow_jury_to_participate": round.AllowJuryToParticipate,
		"allow_multiple_judgement":  round.AllowMultipleJudgement,
		"secret_ballot":             round.SecretBallot,
		"article_allow_expansions":  round.ArticleAllowExpansions,
		"article_allow_creations":   round.ArticleAllowCreations,
	}
	if err := conn.Model(round).Updates(booleans).Error; err != nil {
		return nil, err
	}
	return round, nil
}
func (r *RoundRepository) FindAll(conn *gorm.DB, filter *models.RoundFilter) ([]models.Round, error) {
	var rounds []models.Round
	where := &models.Round{}
	if filter != nil {
		if filter.CampaignID != "" {
			where.CampaignID = filter.CampaignID
		}
	}
	stmt := conn.Where(where)
	if filter.Limit > 0 {
		stmt = stmt.Limit(filter.Limit)
	}
	result := stmt.Find(&rounds)
	return rounds, result.Error
}
func (r *RoundRepository) FindByID(conn *gorm.DB, id models.IDType) (*models.Round, error) {
	round := &models.Round{}
	where := &models.Round{RoundID: models.IDType(id)}
	result := conn.First(round, where)
	return round, result.Error
}
func (r *RoundRepository) GetResults(conn *gorm.DB, roundID models.IDType, qry *models.SubmissionResultQuery) (results []models.SubmissionResult, err error) {
	results = []models.SubmissionResult{}
	q := query.Use(conn)
	Submission := q.Submission
	stmt := Submission.Select(Submission.SubmissionID, Submission.Name, Submission.Score, Submission.Author, Submission.EvaluationCount, Submission.MediaType).
		Where(Submission.RoundID.Eq(roundID.String()))
	if qry != nil {
		if qry.Limit > 0 {
			stmt = stmt.Limit(qry.Limit)
		}
		if qry.ContinueToken != "" {
			stmt = stmt.Where(Submission.SubmissionID.Gt(qry.ContinueToken))
		}
		if qry.PreviousToken != "" {
			stmt = stmt.Where(Submission.SubmissionID.Lt(qry.PreviousToken))
		}
		if len(qry.Type) > 0 {
			types := []string{}
			for _, t := range qry.Type {
				types = append(types, string(t))
			}
			stmt = stmt.Where(Submission.MediaType.In(types...))
		}
	}

	err = stmt.Order(Submission.Score.Desc(), Submission.EvaluationCount.Desc()).Scan(&results)
	if err != nil {
		return nil, err
	}
	return results, nil

}
func (r *RoundRepository) GetResultSummary(conn *gorm.DB, roundID models.IDType) (results []models.EvaluationResult, err error) {
	results = []models.EvaluationResult{}
	q := query.Use(conn)
	stmt := q.Submission.Select(q.Submission.Score.As("AverageScore"), q.Submission.SubmissionID.Count().As("SubmissionCount")).
		Where(q.Submission.RoundID.Eq(roundID.String())).
		Group(q.Submission.Score).Order(q.Submission.Score.Desc()).Limit(100)
	err = stmt.Scan(&results)
	return results, err

}
func (r *RoundRepository) Delete(conn *gorm.DB, id models.IDType) error {
	round := &models.Round{RoundID: id}
	result := conn.Delete(round)
	return result.Error
}
func (r *RoundRepository) UpdateStatisticsByRoundID(conn *gorm.DB, roundID models.IDType) error {
	q := query.Use(conn)
	err := q.RoundStatistics.UpdateByRoundID(roundID.String())
	if err != nil {
		return err
	}
	return q.JuryStatistics.TriggerByRoundID(roundID.String())
}
func (r *RoundRepository) UpdateFullStatisticsByRoundID(conn *gorm.DB, roundID models.IDType) error {
	q := query.Use(conn)
	err := q.RoundStatistics.UpdateByRoundID(roundID.String())
	if err != nil {
		return err
	}
	return q.SubmissionStatistics.TriggerByRoundId(roundID.String())
}
