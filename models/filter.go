package models

type CommonFilter struct {
	Limit         int    `form:"limit"`
	ContinueToken string `form:"next"`
	PreviousToken string `form:"prev"`
}
