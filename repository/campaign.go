package repository

import (
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"strings"

	"gorm.io/gorm"
)

type CampaignRepository struct{}

func NewCampaignRepository() *CampaignRepository {
	return &CampaignRepository{}
}
func (c *CampaignRepository) Create(conn *gorm.DB, campaign *models.Campaign) error {
	result := conn.Create(campaign)
	return result.Error
}
func (c *CampaignRepository) FindByID(conn *gorm.DB, id models.IDType) (*models.Campaign, error) {
	campaign := &models.Campaign{}
	q := query.Use(conn)
	campaign, err := q.Campaign.Where(q.Campaign.CampaignID.Eq(id.String())).First()
	return campaign, err
}
func (c *CampaignRepository) ListAllCampaigns(conn *gorm.DB, qry *models.CampaignFilter) ([]models.Campaign, error) {
	var campaigns []models.Campaign
	q := query.Use(conn)
	Campaign := q.Campaign
	stmt := Campaign.Select(Campaign.ALL)
	if qry != nil {
		if qry.Limit > 0 {
			stmt = stmt.Limit(qry.Limit)
		}
		if len(qry.IDs) > 0 {
			idCopies := []string{}
			for _, id := range qry.IDs {
				if id != "" && strings.Contains(string(id), ",") {
					idCopies = append(idCopies, strings.Split(string(id), ",")...)
				} else {
					idCopies = append(idCopies, string(id))
				}
			}
			stmt = stmt.Where(Campaign.CampaignID.In(idCopies...))
		}
		if qry.IsClosed != nil {
			if *qry.IsClosed {
				stmt = stmt.Unscoped().Where(Campaign.ArchivedAt.IsNotNull())
			} else {
				stmt = stmt.Where(Campaign.ArchivedAt.IsNull())
			}
		}
		if qry.IsHidden != nil {
			if *qry.IsHidden {
				stmt = stmt.Where(Campaign.IsPublic.Not())
			} else {
				stmt = stmt.Where(Campaign.IsPublic)
			}

		}
		if qry.ProjectID != "" {
			stmt = stmt.Where(Campaign.ProjectID.Eq(qry.ProjectID.String()))
		}
		if qry.SortOrder == models.SortOrderAsc {
			stmt = stmt.Order(q.Campaign.CampaignID.Asc())
		} else {
			stmt = stmt.Order(q.Campaign.CampaignID.Desc())
		}
	}
	err := stmt.Scan(&campaigns)
	return campaigns, err
}
func (c *CampaignRepository) Update(conn *gorm.DB, campaign *models.Campaign) error {
	result := conn.Updates(campaign)
	return result.Error
}
func (c *CampaignRepository) UpdateLatestRound(tx *gorm.DB, campaignID models.IDType) error {
	q := query.Use(tx)
	Campaign := q.Campaign
	Round := q.Round
	latestRound := Round.Select(Round.RoundID).Where(Round.CampaignID.Eq(campaignID.String())).Order(Round.RoundID.Desc()).Limit(1)
	result, err := Campaign.Where(Campaign.CampaignID.Eq(campaignID.String())).Update(Campaign.LatestRoundID, latestRound)
	if err != nil {
		return err
	}
	return result.Error
}
func (c *CampaignRepository) ArchiveCampaign(conn *gorm.DB, campaignID models.IDType) error {
	result := conn.Delete(&models.Campaign{}, campaignID)
	return result.Error
}
func (c *CampaignRepository) UnArchiveCampaign(conn *gorm.DB, campaignID models.IDType) error {
	q := query.Use(conn.Unscoped())
	Campaign := q.Campaign
	campaign, err := Campaign.Where(Campaign.CampaignID.Eq(campaignID.String())).Update(Campaign.ArchivedAt, nil)
	if err != nil {
		return err
	}
	if campaign.Error != nil {
		return campaign.Error
	}
	return nil
}
