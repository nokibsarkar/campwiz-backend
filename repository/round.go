package repository

import (
	"nokib/campwiz/models"

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
	result := conn.Save(round)
	if result.Error != nil {
		return nil, result.Error
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
func (r *RoundRepository) GetResults(conn *gorm.DB, roundID models.IDType) (results []models.EvaluationResult, err error) {
	results = []models.EvaluationResult{}
	stmt := conn.Model(&models.Submission{}).Select("score as AverageScore, count(submission_id) as SubmissionCount").Where(&models.Submission{CurrentRoundID: roundID}).Group("score").Order("score desc").Find(&results)
	if stmt.Error != nil {
		return nil, stmt.Error
	}
	return results, nil

}
