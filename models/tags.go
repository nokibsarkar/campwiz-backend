package models

type Tag struct {
	Name       string   `gorm:"primaryKey;type:varchar(255);not null" json:"name"`
	CampaignId int      `gorm:"primaryKey;not null" json:"campaignId"`
	Campaign   Campaign `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}
