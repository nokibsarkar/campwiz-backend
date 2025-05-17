package distributionstrategy

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
