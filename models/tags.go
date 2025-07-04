package models

type Tag struct {
	Name       string    `gorm:"primaryKey;type:varchar(255);not null" json:"name"`
	CampaignID IDType    `gorm:"primaryKey;not null;type:varchar(255);collate:utf8mb4_general_ci" json:"campaignId"`
	Campaign   *Campaign `gorm:"foreignKey:CampaignID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}
