package models

// This represents the username provided by wikimedia, it should be normalized already by wikimedia
type WikimediaUsernameType string
type UserIDType string
type CampaignIDType string
type RoundIDType string
type TaskIDType string

// type SubmissionIDType string
type RoleIDType IDType
type EvaluationIDType string
type IDType string
type ScoreType float64

func (i *IDType) String() string {
	return string(*i)
}
func (i *IDType) GormDataType() string {
	return "varchar(255)"
}

func (i *UserIDType) String() string {
	return string(*i)
}
func (i *ScoreType) GormDataType() string {
	return "Decimal(5,2)"
}

// func (i *ScoreType) Scan(value int64) error {
// 	log.Printf("Scanning score: %v\n", value)
// 	*i = ScoreType(value / 100.0)
// 	return errors.New("Not implemented")
// }
// func (i *ScoreType) Value() (int64, error) {
// 	log.Printf("Value of score: %v\n", *i)
// 	return int64(*i * 100), nil
// }
