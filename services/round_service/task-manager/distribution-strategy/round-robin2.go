package distributionstrategy

// Prevent Self evaluation SQL: update evaluations u1 join (select judge_id, evaluation_id, name from evaluations join submissions join roles on evaluations.submission_id = submissions.submission_id and evaluations.judge_id = roles.role_id  where submitted_by_id=roles.user_id and evaluations.round_id='r2eczdvrjl2ps') u2 using(evaluation_id) set u1.judge_id = (select role_id from roles where role_id <> u2.judge_id and round_id='r2eczdvrjl2ps' and role_id not in (select judge_id from evaluations where submission_id = u1.submission_id) order by rand() limit 1) where round_id='r2eczdvrjl2ps' and score is null;
// This method would distribute all the evaluations to the juries in round robin fashion
func (strategy *RoundRobinDistributionStrategy) AssignJuries2() {
	// From these jury, we would be collecting the evaluations
	// sourceRoleIds := []string{}
	// roundId := ""
	// // These jury would be assigned to the evaluations
	// targetRoleIds := []string{}
	// newAssignment := map[string]int{}
	// /*
	// 	Assignment would happen in the following way:
	// 		- Any image that is submitted by myself cannot be evaluated by me
	// 		- Any submission that was assigned to me by a previous assignment
	// 		cannot be assigned to me again with a different submission
	// */
	// q, err := repository.GetDBWithGen()
	// if err != nil {
	// 	return
	// }
	// Evaluation := q.Evaluation
	// for _, juryId := range targetRoleIds {
	// 	// create a list of potential juries whose evaluations I can be assigned to
	// 	canBeAssigned := newAssignment[juryId]
	// 	if canBeAssigned == 0 {
	// 		continue
	// 	}
	// 	// assign me to the evaluation
	// 	Evaluation.Update()
	// }
}
