package consts

import (
	"database/sql/driver"
)

// Permission is a type to represent a permission
type Permission uint64

// PermissionGroup is a type to represent a group of permissions
type PermissionGroup Permission

type UintString string
type PermissionName string

const (
	// System level permissions
	// Permission to login (it would be removed for the banned users)
	PermissionLogin Permission = 1 << iota

	// Permission to see the list of users
	PermissionSeeAllUsers
	// Permission to see the details of a user
	PermissionSeeUserDetails

	// Project Level Permissions (for the admin of the system)
	PermissionCreateProject // Permission to Create a Project
	PermissionUpdateProject // Permission to Update an existing Project, like adding lead, changing name,
	PermissionDeleteProject // Permission to Delete a Project
	// the following permission would mean the person has access to all the projects
	PermissionOtherProjectAccess // Permission to access projects other than the one the user is a lead of

	// Campaign Level Permissions (for the lead of the project)
	PermissionCreateCampaign // Permission to Create a Campaign
	// Permission to Update an existing Campaign (but it's not possible to change the status of the campaign)
	PermissionUpdateCampaignDetails
	// Permission to Change the status of a Campaign, but not to update the details of the campaign
	PermissionUpdateCampaignStatus
	// Permission to Delete an existing Campaign
	PermissionDeleteCampaign
	// Permission to Approve or Reject a Campaign (should be held by the lead of the project)
	// campaigns can be created by co-ordinators, but they need to be approved by the lead
	PermissionApproveRejectCampaign

	// Round Level Permissions (for the coordinator of the campaign)

	// Permission to Create a Round on an existing Campaign
	PermissionCreateRound
	// Permission to update the details of a Round
	PermissionUpdateRoundDetails
	// Permission to update the status of a Round
	PermissionUpdateRoundStatus
	// Permission to Delete a Round
	PermissionDeleteRound
	// Permission to See self Evaluation Result
	PermissionSeeOwnEvaluationResult
	// Permission to See Evaluation Result of others
	PermissionSeeOthersEvaluationResult

	// Permission to evaluate a submission
	PermissionEvaluateSubmission
	// Permission to submit a submission
	PermissionSubmitSubmission
	PermissionRandomize
)
const (
	// PermissionGroups
	PermissionGroupBanned = PermissionGroup(0)
	PermissionGroupUSER   = PermissionGroup(PermissionLogin)
	/*
		A jury can:
			- see all the submissions
			- evaluate the submissions
			- see the evaluation results of the evaluations they have done
	*/
	PermissionGroupJury = (PermissionGroupUSER | // Get all the permissions of a user
		PermissionGroup(PermissionEvaluateSubmission) | // evaluate the submissions
		PermissionGroup(PermissionSeeOwnEvaluationResult)) // see the evaluation results of the evaluations they have done
	/*
		A coordinator can:
			- see all the submissions
			- See the evaluation results of the evaluations others have done (including he himself)
			- create a campaign (but it needs to be approved by the lead)
			- create a round (once the campaign is approved by the lead)
			- update the details of a round (like start time, end time, public/private, etc)
			- update the status of a round (like open, closed, etc)
			- delete a round , only if any one of the conditions is met:
				- the round is not started yet
				- the round is started but no submissions have been made
				- the round has submissions but no other round depends on this round
	*/
	PermissionGroupCoordinator = (PermissionGroupUSER | // Get all the permissions of a user
		PermissionGroup(PermissionCreateCampaign) | // create a campaign (but it needs to be approved by the lead)
		PermissionGroup(PermissionCreateRound) | // create a round (once the campaign is approved by the lead)
		PermissionGroup(PermissionUpdateRoundDetails) | // update the details of a round (like start time, end time, public/private, etc)
		PermissionGroup(PermissionUpdateRoundStatus) | // update the status of a round (like open, closed, etc)
		PermissionGroup(PermissionDeleteRound) | // delete a round , only if any one of the conditions is met:
		PermissionGroup(PermissionSeeOthersEvaluationResult) | // See the evaluation results of the evaluations others have done (including he himself)
		PermissionGroup(PermissionSeeOwnEvaluationResult)) | // see the evaluation results of the evaluations they have done
		PermissionGroup(PermissionRandomize) // randomize the submissions
	/*
		A lead can:
			- Get all the permissions of a coordinator
			- approve or reject a campaign
			- update the details of a campaign
			- update the status of a campaign
			- delete a campaign
	*/
	PermissionGroupLead = (PermissionGroupCoordinator | // Get all the permissions of a coordinator
		PermissionGroup(PermissionApproveRejectCampaign) | // approve or reject a campaign (created by a coordinator)
		PermissionGroup(PermissionUpdateCampaignDetails) | // update the details of a campaign
		PermissionGroup(PermissionUpdateCampaignStatus) | // update the status of a campaign
		PermissionGroup(PermissionDeleteCampaign)) // delete a campaign
	/*
		An admin can:
			- Get all the permissions of a lead
			- create a project
			- update the details of a project
			- delete a project
			- access other projects
	*/
	PermissionGroupADMIN = (PermissionGroupLead | // Get all the permissions of a lead
		PermissionGroup(PermissionCreateProject) | // create a project
		PermissionGroup(PermissionUpdateProject) | // update the details of a project
		PermissionGroup(PermissionDeleteProject) | // delete a project
		PermissionGroup(PermissionOtherProjectAccess)) // access other projects
)

// HasPermission checks if a permission has a permission
func (pg PermissionGroup) HasPermission(p Permission) bool {
	return pg&PermissionGroup(p) == PermissionGroup(p)
}

// AddPermission adds a permission to a permission group
func (pg *PermissionGroup) AddPermission(p Permission) {
	*pg |= PermissionGroup(p)
}

// RemovePermission removes a permission from a permission group
func (pg *PermissionGroup) RemovePermission(p Permission) {
	*pg &= ^PermissionGroup(p)
}
func (p *PermissionGroup) Scan(value interface{}) error {
	*p = PermissionGroup(value.(int64))
	return nil

}

func (p *PermissionGroup) Value() (driver.Value, error) {
	return int64(*p), nil
}

type PermissionMap map[PermissionName]Permission

func GetPermissionMap() PermissionMap {
	permissionMap := PermissionMap{}
	permissionMap[PermissionName("PermissionApproveRejectCampaign")] = PermissionApproveRejectCampaign
	permissionMap[PermissionName("PermissionCreateCampaign")] = PermissionCreateCampaign
	permissionMap[PermissionName("PermissionCreateProject")] = PermissionCreateProject
	permissionMap[PermissionName("PermissionCreateRound")] = PermissionCreateRound
	permissionMap[PermissionName("PermissionDeleteCampaign")] = PermissionDeleteCampaign
	permissionMap[PermissionName("PermissionDeleteProject")] = PermissionDeleteProject
	permissionMap[PermissionName("PermissionDeleteRound")] = PermissionDeleteRound
	permissionMap[PermissionName("PermissionEvaluateSubmission")] = PermissionEvaluateSubmission
	permissionMap[PermissionName("PermissionLogin")] = PermissionLogin
	permissionMap[PermissionName("PermissionOtherProjectAccess")] = PermissionOtherProjectAccess
	permissionMap[PermissionName("PermissionSeeAllUsers")] = PermissionSeeAllUsers
	permissionMap[PermissionName("PermissionSeeOthersEvaluationResult")] = PermissionSeeOthersEvaluationResult
	permissionMap[PermissionName("PermissionSeeOwnEvaluationResult")] = PermissionSeeOwnEvaluationResult
	permissionMap[PermissionName("PermissionSeeUserDetails")] = PermissionSeeUserDetails
	permissionMap[PermissionName("PermissionSubmitSubmission")] = PermissionSubmitSubmission
	permissionMap[PermissionName("PermissionUpdateCampaignDetails")] = PermissionUpdateCampaignDetails
	permissionMap[PermissionName("PermissionUpdateCampaignStatus")] = PermissionUpdateCampaignStatus
	permissionMap[PermissionName("PermissionUpdateProject")] = PermissionUpdateProject
	permissionMap[PermissionName("PermissionUpdateRoundDetails")] = PermissionUpdateRoundDetails
	permissionMap[PermissionName("PermissionUpdateRoundStatus")] = PermissionUpdateRoundStatus
	permissionMap[PermissionName("PermissionSeeOthersEvaluationResult")] = PermissionSeeOthersEvaluationResult
	permissionMap[PermissionName("PermissionSeeOwnEvaluationResult")] = PermissionSeeOwnEvaluationResult
	permissionMap[PermissionName("PermissionEvaluateSubmission")] = PermissionEvaluateSubmission
	permissionMap[PermissionName("PermissionSubmitSubmission")] = PermissionSubmitSubmission
	return permissionMap
}
func (p *PermissionGroup) GetPermissions(permMap PermissionMap) []PermissionName {
	permissions := []PermissionName{}
	for pName, pInt := range permMap {
		if p.HasPermission(pInt) {
			permissions = append(permissions, pName)
		}
	}
	return permissions
}
