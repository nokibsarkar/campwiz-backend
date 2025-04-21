package main

import (
	"nokib/campwiz/models"
	"nokib/campwiz/repository/cache"

	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../query",                                                         // output path
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})
	g.ApplyBasic(models.Project{}, models.User{}, models.Campaign{},
		models.Round{}, models.Task{}, models.Role{}, models.Submission{},
		models.Evaluation{}, cache.Evaluation{}, models.SubmissionResult{}, models.TaskData{}, models.Category{})
	g.ApplyInterface(func(cache.Dirtributor) {}, cache.Evaluation{})
	g.ApplyInterface(func(models.SubmissionStatisticsFetcher) {}, models.SubmissionStatistics{})
	g.ApplyInterface(func(models.JuryStatisticsUpdater) {}, models.JuryStatistics{})
	g.ApplyInterface(func(models.RoundStatisticsFetcher) {}, models.RoundStatistics{})
	g.ApplyInterface(func(models.Evaluator) {}, models.Evaluation{})
	g.ApplyInterface(func(models.SubmissionFetcher) {}, models.CommonsSubmissionEntry{})
	g.Execute()
}
