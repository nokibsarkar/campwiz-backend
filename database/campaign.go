package database

import (
	"nokib/campwiz/consts"
	"strings"
	"time"

	"gorm.io/gorm"
)

type CampaignWithWriteableFields struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	StartDate   time.Time       `json:"startDate"`
	EndDate     time.Time       `json:"endDate"`
	Language    consts.Language `json:"language"`
	Rules       string          `json:"rules"`
	Image       string          `json:"image"`
	Roles       []Role          `json:"roles"`
}
type Campaign struct {
	CampaignID IDType `gorm:"primaryKey" json:"campaignId"`
	// read only
	CreatedAt   *time.Time `json:"createdAt" gorm:"-<-:create"`
	CreatedByID IDType     `json:"createdById" gorm:"index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CampaignWithWriteableFields
	CreatedBy *User `json:"-" gorm:"foreignKey:CreatedByID"`
}
type CampaignFilter struct {
	IDs []IDType `form:"ids,omitEmpty"`
	CommonFilter
}
type CampaignRepository struct{}

func NewCampaignRepository() *CampaignRepository {
	return &CampaignRepository{}
}
func (c *CampaignRepository) Create(conn *gorm.DB, campaign *Campaign) error {
	result := conn.Create(campaign)
	return result.Error
}
func (c *CampaignRepository) FindByID(conn *gorm.DB, id IDType) (*Campaign, error) {
	campaign := &Campaign{}
	where := &Campaign{CampaignID: id}
	result := conn.First(campaign, where)
	return campaign, result.Error
}
func (c *CampaignRepository) ListAllCampaigns(conn *gorm.DB, query *CampaignFilter) ([]Campaign, error) {
	var campaigns []Campaign
	stmt := conn
	if query != nil {
		if query.Limit > 0 {
			stmt = stmt.Limit(query.Limit)
		}
		if len(query.IDs) > 0 {
			idCopies := []string{}
			for _, id := range query.IDs {
				if id != "" && strings.Contains(string(id), ",") {
					idCopies = append(idCopies, strings.Split(string(id), ",")...)
				} else {
					idCopies = append(idCopies, string(id))
				}
			}
			stmt = stmt.Where("id IN (?)", idCopies)
		}
	}
	result := stmt.Find(&campaigns)
	return campaigns, result.Error
}
func (c *CampaignRepository) Update(conn *gorm.DB, campaign *Campaign) error {
	result := conn.Save(campaign)
	return result.Error
}
