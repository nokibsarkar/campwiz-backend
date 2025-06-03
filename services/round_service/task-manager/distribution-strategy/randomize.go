package distributionstrategy

import (
	"context"
	"errors"
	"log"
	"nokib/campwiz/models"
	"nokib/campwiz/query"
	"nokib/campwiz/repository"

	"gorm.io/gorm"
)

type Randomizer struct{}

func (d *DistributorServer) Randomize(ctx context.Context, req *models.DistributeWithRoundRobinRequest) (*models.DistributeWithRoundRobinResponse, error) {
	roundId := models.IDType(req.RoundId)
	go randomize(ctx, roundId) //nolint:errcheck
	return &models.DistributeWithRoundRobinResponse{
		TaskId: req.TaskId,
	}, nil
}
func randomize(ctx context.Context, roundId models.IDType) error {
	conn, close, err := repository.GetDB(ctx)
	if err != nil {
		return err
	}
	defer close()
	round_repo := repository.NewRoundRepository()
	roleRepo := repository.NewRoleRepository()
	round, err := round_repo.FindByID(conn, roundId)
	if err != nil {
		log.Println("Error finding round:", err)
		return err
	}
	if round == nil {
		log.Println("Round not found")
		return errors.New("round not found")
	}
	juryType := models.RoleTypeJury
	roles, err := roleRepo.ListAllRoles(conn, &models.RoleFilter{
		RoundID: &round.RoundID,
		Type:    &juryType,
	})
	if err != nil {
		log.Println("Error getting roles:", err)
		return err
	}
	if len(roles) == 0 {
		log.Println("No roles found")
		return errors.New("no roles found")
	}
	q := query.Use(conn)
	Evaluation := q.Evaluation

	for _, role := range roles {

		log.Println("Randomized submissions for role:", role.RoleID)
		unevaluatedEvaluations, err := Evaluation.Where(Evaluation.RoundID.Eq(round.RoundID.String()), Evaluation.JudgeID.Eq(role.RoleID.String()), Evaluation.Score.IsNull()).Find()
		if err != nil {
			log.Println("Error getting evaluations:", err)
			continue
		}

		swappableLimit := len(unevaluatedEvaluations)
		if swappableLimit == 0 {
			log.Println("No unevaluated evaluations found for role:", role.RoleID)
			continue
		}
		targetSwappables, err := Evaluation.FetchTargetSwappables(string(roundId), role.RoleID.String(), swappableLimit)
		// Execute the statement or log it to ensure it's used
		if err != nil {
			log.Println("Error executing query:", err)
		}
		swappableLimit = min(len(targetSwappables), swappableLimit)
		if swappableLimit == 0 {
			log.Println("Cannot find any target swappable for role:", role.RoleID)
			continue
		}
		//dutoke soman korte hobe
		targetSwappables = targetSwappables[:swappableLimit]
		unevaluatedEvaluations = unevaluatedEvaluations[:swappableLimit]
		ami := role.RoleID.String()
		denaPawna := map[string][]string{}
		metanorLimit := 0
		const maxLimit = 5000
		for cursor := range swappableLimit {
			// targeter jeta ami swap korte chai
			amiNebo := targetSwappables[cursor].EvaluationID.String()
			// amar name register kori
			denaPawna[ami] = append(denaPawna[ami], amiNebo)
			// amar kach theke kake dite chai
			kakeDebo := targetSwappables[cursor].JudgeID.String()
			// ki debo take
			kiDebo := unevaluatedEvaluations[cursor].EvaluationID.String()
			denaPawna[kakeDebo] = append(denaPawna[kakeDebo], kiDebo)
			metanorLimit++
			if metanorLimit >= maxLimit {
				log.Printf("Randomizing %v submissions for role %s", metanorLimit, role.RoleID)
				if err := mitiyeNao(conn, denaPawna); err != nil {
					log.Println("Error updating evaluations:", err)
					return err
				}
				denaPawna = map[string][]string{}
				metanorLimit = 0
			}

		}
		if metanorLimit > 0 {
			if err := mitiyeNao(conn, denaPawna); err != nil {
				log.Println("Error updating evaluations:", err)
				return err
			}
		}
	}
	return nil
}

func mitiyeNao(conn *gorm.DB, amarDena map[string][]string) error {
	// Update the evaluations in the database
	tx := conn.Begin()
	q := query.Use(tx)

	Evaluation := q.Evaluation
	for judgeId, submissionIds := range amarDena {
		res, err := Evaluation.Where(Evaluation.EvaluationID.In(submissionIds...)).Limit(len(submissionIds)).Update(Evaluation.JudgeID, judgeId)
		if err != nil {
			log.Println("Error updating evaluations:", err)
			tx.Rollback()
			return err
		}
		if res.RowsAffected == 0 {
			log.Println("No rows affected for judge ID:", judgeId)
			continue
		}
		log.Printf("Updated %d evaluations for judge ID %s", res.RowsAffected, judgeId)
	}
	tx.Commit()
	return nil
}
