// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"nokib/campwiz/models"
)

func newSubmission(db *gorm.DB, opts ...gen.DOOption) submission {
	_submission := submission{}

	_submission.submissionDo.UseDB(db, opts...)
	_submission.submissionDo.UseModel(&models.Submission{})

	tableName := _submission.submissionDo.TableName()
	_submission.ALL = field.NewAsterisk(tableName)
	_submission.SubmissionID = field.NewString(tableName, "submission_id")
	_submission.Name = field.NewString(tableName, "name")
	_submission.CampaignID = field.NewString(tableName, "campaign_id")
	_submission.URL = field.NewString(tableName, "url")
	_submission.PageID = field.NewUint64(tableName, "page_id")
	_submission.Score = field.NewFloat64(tableName, "score")
	_submission.Author = field.NewString(tableName, "author")
	_submission.SubmittedByID = field.NewString(tableName, "submitted_by_id")
	_submission.ParticipantID = field.NewString(tableName, "participant_id")
	_submission.RoundID = field.NewString(tableName, "round_id")
	_submission.SubmittedAt = field.NewTime(tableName, "submitted_at")
	_submission.CreatedAtExternal = field.NewTime(tableName, "created_at_external")
	_submission.DistributionTaskID = field.NewString(tableName, "distribution_task_id")
	_submission.ImportTaskID = field.NewString(tableName, "import_task_id")
	_submission.AssignmentCount = field.NewUint(tableName, "assignment_count")
	_submission.EvaluationCount = field.NewUint(tableName, "evaluation_count")
	_submission.MediaType = field.NewString(tableName, "media_type")
	_submission.ThumbURL = field.NewString(tableName, "thumb_url")
	_submission.ThumbWidth = field.NewUint64(tableName, "thumb_width")
	_submission.ThumbHeight = field.NewUint64(tableName, "thumb_height")
	_submission.License = field.NewString(tableName, "license")
	_submission.Description = field.NewString(tableName, "description")
	_submission.CreditHTML = field.NewString(tableName, "credit_html")
	_submission.Metadata = field.NewField(tableName, "metadata")
	_submission.Width = field.NewUint64(tableName, "width")
	_submission.Height = field.NewUint64(tableName, "height")
	_submission.Resolution = field.NewUint64(tableName, "resolution")
	_submission.Duration = field.NewUint64(tableName, "duration")
	_submission.Bitrate = field.NewUint64(tableName, "bitrate")
	_submission.Size = field.NewUint64(tableName, "size")
	_submission.Participant = submissionBelongsToParticipant{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Participant", "models.User"),
		LeadingProject: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Participant.LeadingProject", "models.Project"),
		},
	}

	_submission.Submitter = submissionBelongsToSubmitter{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Submitter", "models.User"),
	}

	_submission.Campaign = submissionBelongsToCampaign{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Campaign", "models.Campaign"),
		CreatedBy: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Campaign.CreatedBy", "models.User"),
		},
		Project: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Campaign.Project", "models.Project"),
		},
		LatestRound: struct {
			field.RelationField
			Campaign struct {
				field.RelationField
			}
			Creator struct {
				field.RelationField
			}
			DependsOnRound struct {
				field.RelationField
			}
			Roles struct {
				field.RelationField
				Round struct {
					field.RelationField
				}
				Campaign struct {
					field.RelationField
				}
				User struct {
					field.RelationField
				}
				Project struct {
					field.RelationField
				}
			}
		}{
			RelationField: field.NewRelation("Campaign.LatestRound", "models.Round"),
			Campaign: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Campaign.LatestRound.Campaign", "models.Campaign"),
			},
			Creator: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Campaign.LatestRound.Creator", "models.User"),
			},
			DependsOnRound: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Campaign.LatestRound.DependsOnRound", "models.Round"),
			},
			Roles: struct {
				field.RelationField
				Round struct {
					field.RelationField
				}
				Campaign struct {
					field.RelationField
				}
				User struct {
					field.RelationField
				}
				Project struct {
					field.RelationField
				}
			}{
				RelationField: field.NewRelation("Campaign.LatestRound.Roles", "models.Role"),
				Round: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Campaign.LatestRound.Roles.Round", "models.Round"),
				},
				Campaign: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Campaign.LatestRound.Roles.Campaign", "models.Campaign"),
				},
				User: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Campaign.LatestRound.Roles.User", "models.User"),
				},
				Project: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Campaign.LatestRound.Roles.Project", "models.Project"),
				},
			},
		},
		CampaignTags: struct {
			field.RelationField
			Campaign struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Campaign.CampaignTags", "models.Tag"),
			Campaign: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Campaign.CampaignTags.Campaign", "models.Campaign"),
			},
		},
		Roles: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Campaign.Roles", "models.Role"),
		},
		Rounds: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Campaign.Rounds", "models.Round"),
		},
	}

	_submission.Round = submissionBelongsToRound{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Round", "models.Round"),
	}

	_submission.DistributionTask = submissionBelongsToDistributionTask{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("DistributionTask", "models.Task"),
		Submittor: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("DistributionTask.Submittor", "models.User"),
		},
		TaskData: struct {
			field.RelationField
			Task struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("DistributionTask.TaskData", "models.TaskData"),
			Task: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("DistributionTask.TaskData.Task", "models.Task"),
			},
		},
	}

	_submission.ImportTask = submissionBelongsToImportTask{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("ImportTask", "models.Task"),
	}

	_submission.fillFieldMap()

	return _submission
}

type submission struct {
	submissionDo

	ALL                field.Asterisk
	SubmissionID       field.String
	Name               field.String
	CampaignID         field.String
	URL                field.String
	PageID             field.Uint64
	Score              field.Float64
	Author             field.String
	SubmittedByID      field.String
	ParticipantID      field.String
	RoundID            field.String
	SubmittedAt        field.Time
	CreatedAtExternal  field.Time
	DistributionTaskID field.String
	ImportTaskID       field.String
	AssignmentCount    field.Uint
	EvaluationCount    field.Uint
	MediaType          field.String
	ThumbURL           field.String
	ThumbWidth         field.Uint64
	ThumbHeight        field.Uint64
	License            field.String
	Description        field.String
	CreditHTML         field.String
	Metadata           field.Field
	Width              field.Uint64
	Height             field.Uint64
	Resolution         field.Uint64
	Duration           field.Uint64
	Bitrate            field.Uint64
	Size               field.Uint64
	Participant        submissionBelongsToParticipant

	Submitter submissionBelongsToSubmitter

	Campaign submissionBelongsToCampaign

	Round submissionBelongsToRound

	DistributionTask submissionBelongsToDistributionTask

	ImportTask submissionBelongsToImportTask

	fieldMap map[string]field.Expr
}

func (s submission) Table(newTableName string) *submission {
	s.submissionDo.UseTable(newTableName)
	return s.updateTableName(newTableName)
}

func (s submission) As(alias string) *submission {
	s.submissionDo.DO = *(s.submissionDo.As(alias).(*gen.DO))
	return s.updateTableName(alias)
}

func (s *submission) updateTableName(table string) *submission {
	s.ALL = field.NewAsterisk(table)
	s.SubmissionID = field.NewString(table, "submission_id")
	s.Name = field.NewString(table, "name")
	s.CampaignID = field.NewString(table, "campaign_id")
	s.URL = field.NewString(table, "url")
	s.PageID = field.NewUint64(table, "page_id")
	s.Score = field.NewFloat64(table, "score")
	s.Author = field.NewString(table, "author")
	s.SubmittedByID = field.NewString(table, "submitted_by_id")
	s.ParticipantID = field.NewString(table, "participant_id")
	s.RoundID = field.NewString(table, "round_id")
	s.SubmittedAt = field.NewTime(table, "submitted_at")
	s.CreatedAtExternal = field.NewTime(table, "created_at_external")
	s.DistributionTaskID = field.NewString(table, "distribution_task_id")
	s.ImportTaskID = field.NewString(table, "import_task_id")
	s.AssignmentCount = field.NewUint(table, "assignment_count")
	s.EvaluationCount = field.NewUint(table, "evaluation_count")
	s.MediaType = field.NewString(table, "media_type")
	s.ThumbURL = field.NewString(table, "thumb_url")
	s.ThumbWidth = field.NewUint64(table, "thumb_width")
	s.ThumbHeight = field.NewUint64(table, "thumb_height")
	s.License = field.NewString(table, "license")
	s.Description = field.NewString(table, "description")
	s.CreditHTML = field.NewString(table, "credit_html")
	s.Metadata = field.NewField(table, "metadata")
	s.Width = field.NewUint64(table, "width")
	s.Height = field.NewUint64(table, "height")
	s.Resolution = field.NewUint64(table, "resolution")
	s.Duration = field.NewUint64(table, "duration")
	s.Bitrate = field.NewUint64(table, "bitrate")
	s.Size = field.NewUint64(table, "size")

	s.fillFieldMap()

	return s
}

func (s *submission) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := s.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (s *submission) fillFieldMap() {
	s.fieldMap = make(map[string]field.Expr, 36)
	s.fieldMap["submission_id"] = s.SubmissionID
	s.fieldMap["name"] = s.Name
	s.fieldMap["campaign_id"] = s.CampaignID
	s.fieldMap["url"] = s.URL
	s.fieldMap["page_id"] = s.PageID
	s.fieldMap["score"] = s.Score
	s.fieldMap["author"] = s.Author
	s.fieldMap["submitted_by_id"] = s.SubmittedByID
	s.fieldMap["participant_id"] = s.ParticipantID
	s.fieldMap["round_id"] = s.RoundID
	s.fieldMap["submitted_at"] = s.SubmittedAt
	s.fieldMap["created_at_external"] = s.CreatedAtExternal
	s.fieldMap["distribution_task_id"] = s.DistributionTaskID
	s.fieldMap["import_task_id"] = s.ImportTaskID
	s.fieldMap["assignment_count"] = s.AssignmentCount
	s.fieldMap["evaluation_count"] = s.EvaluationCount
	s.fieldMap["media_type"] = s.MediaType
	s.fieldMap["thumb_url"] = s.ThumbURL
	s.fieldMap["thumb_width"] = s.ThumbWidth
	s.fieldMap["thumb_height"] = s.ThumbHeight
	s.fieldMap["license"] = s.License
	s.fieldMap["description"] = s.Description
	s.fieldMap["credit_html"] = s.CreditHTML
	s.fieldMap["metadata"] = s.Metadata
	s.fieldMap["width"] = s.Width
	s.fieldMap["height"] = s.Height
	s.fieldMap["resolution"] = s.Resolution
	s.fieldMap["duration"] = s.Duration
	s.fieldMap["bitrate"] = s.Bitrate
	s.fieldMap["size"] = s.Size

}

func (s submission) clone(db *gorm.DB) submission {
	s.submissionDo.ReplaceConnPool(db.Statement.ConnPool)
	s.Participant.db = db.Session(&gorm.Session{Initialized: true})
	s.Participant.db.Statement.ConnPool = db.Statement.ConnPool
	s.Submitter.db = db.Session(&gorm.Session{Initialized: true})
	s.Submitter.db.Statement.ConnPool = db.Statement.ConnPool
	s.Campaign.db = db.Session(&gorm.Session{Initialized: true})
	s.Campaign.db.Statement.ConnPool = db.Statement.ConnPool
	s.Round.db = db.Session(&gorm.Session{Initialized: true})
	s.Round.db.Statement.ConnPool = db.Statement.ConnPool
	s.DistributionTask.db = db.Session(&gorm.Session{Initialized: true})
	s.DistributionTask.db.Statement.ConnPool = db.Statement.ConnPool
	s.ImportTask.db = db.Session(&gorm.Session{Initialized: true})
	s.ImportTask.db.Statement.ConnPool = db.Statement.ConnPool
	return s
}

func (s submission) replaceDB(db *gorm.DB) submission {
	s.submissionDo.ReplaceDB(db)
	s.Participant.db = db.Session(&gorm.Session{})
	s.Submitter.db = db.Session(&gorm.Session{})
	s.Campaign.db = db.Session(&gorm.Session{})
	s.Round.db = db.Session(&gorm.Session{})
	s.DistributionTask.db = db.Session(&gorm.Session{})
	s.ImportTask.db = db.Session(&gorm.Session{})
	return s
}

type submissionBelongsToParticipant struct {
	db *gorm.DB

	field.RelationField

	LeadingProject struct {
		field.RelationField
	}
}

func (a submissionBelongsToParticipant) Where(conds ...field.Expr) *submissionBelongsToParticipant {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a submissionBelongsToParticipant) WithContext(ctx context.Context) *submissionBelongsToParticipant {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a submissionBelongsToParticipant) Session(session *gorm.Session) *submissionBelongsToParticipant {
	a.db = a.db.Session(session)
	return &a
}

func (a submissionBelongsToParticipant) Model(m *models.Submission) *submissionBelongsToParticipantTx {
	return &submissionBelongsToParticipantTx{a.db.Model(m).Association(a.Name())}
}

func (a submissionBelongsToParticipant) Unscoped() *submissionBelongsToParticipant {
	a.db = a.db.Unscoped()
	return &a
}

type submissionBelongsToParticipantTx struct{ tx *gorm.Association }

func (a submissionBelongsToParticipantTx) Find() (result *models.User, err error) {
	return result, a.tx.Find(&result)
}

func (a submissionBelongsToParticipantTx) Append(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a submissionBelongsToParticipantTx) Replace(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a submissionBelongsToParticipantTx) Delete(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a submissionBelongsToParticipantTx) Clear() error {
	return a.tx.Clear()
}

func (a submissionBelongsToParticipantTx) Count() int64 {
	return a.tx.Count()
}

func (a submissionBelongsToParticipantTx) Unscoped() *submissionBelongsToParticipantTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type submissionBelongsToSubmitter struct {
	db *gorm.DB

	field.RelationField
}

func (a submissionBelongsToSubmitter) Where(conds ...field.Expr) *submissionBelongsToSubmitter {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a submissionBelongsToSubmitter) WithContext(ctx context.Context) *submissionBelongsToSubmitter {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a submissionBelongsToSubmitter) Session(session *gorm.Session) *submissionBelongsToSubmitter {
	a.db = a.db.Session(session)
	return &a
}

func (a submissionBelongsToSubmitter) Model(m *models.Submission) *submissionBelongsToSubmitterTx {
	return &submissionBelongsToSubmitterTx{a.db.Model(m).Association(a.Name())}
}

func (a submissionBelongsToSubmitter) Unscoped() *submissionBelongsToSubmitter {
	a.db = a.db.Unscoped()
	return &a
}

type submissionBelongsToSubmitterTx struct{ tx *gorm.Association }

func (a submissionBelongsToSubmitterTx) Find() (result *models.User, err error) {
	return result, a.tx.Find(&result)
}

func (a submissionBelongsToSubmitterTx) Append(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a submissionBelongsToSubmitterTx) Replace(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a submissionBelongsToSubmitterTx) Delete(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a submissionBelongsToSubmitterTx) Clear() error {
	return a.tx.Clear()
}

func (a submissionBelongsToSubmitterTx) Count() int64 {
	return a.tx.Count()
}

func (a submissionBelongsToSubmitterTx) Unscoped() *submissionBelongsToSubmitterTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type submissionBelongsToCampaign struct {
	db *gorm.DB

	field.RelationField

	CreatedBy struct {
		field.RelationField
	}
	Project struct {
		field.RelationField
	}
	LatestRound struct {
		field.RelationField
		Campaign struct {
			field.RelationField
		}
		Creator struct {
			field.RelationField
		}
		DependsOnRound struct {
			field.RelationField
		}
		Roles struct {
			field.RelationField
			Round struct {
				field.RelationField
			}
			Campaign struct {
				field.RelationField
			}
			User struct {
				field.RelationField
			}
			Project struct {
				field.RelationField
			}
		}
	}
	CampaignTags struct {
		field.RelationField
		Campaign struct {
			field.RelationField
		}
	}
	Roles struct {
		field.RelationField
	}
	Rounds struct {
		field.RelationField
	}
}

func (a submissionBelongsToCampaign) Where(conds ...field.Expr) *submissionBelongsToCampaign {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a submissionBelongsToCampaign) WithContext(ctx context.Context) *submissionBelongsToCampaign {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a submissionBelongsToCampaign) Session(session *gorm.Session) *submissionBelongsToCampaign {
	a.db = a.db.Session(session)
	return &a
}

func (a submissionBelongsToCampaign) Model(m *models.Submission) *submissionBelongsToCampaignTx {
	return &submissionBelongsToCampaignTx{a.db.Model(m).Association(a.Name())}
}

func (a submissionBelongsToCampaign) Unscoped() *submissionBelongsToCampaign {
	a.db = a.db.Unscoped()
	return &a
}

type submissionBelongsToCampaignTx struct{ tx *gorm.Association }

func (a submissionBelongsToCampaignTx) Find() (result *models.Campaign, err error) {
	return result, a.tx.Find(&result)
}

func (a submissionBelongsToCampaignTx) Append(values ...*models.Campaign) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a submissionBelongsToCampaignTx) Replace(values ...*models.Campaign) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a submissionBelongsToCampaignTx) Delete(values ...*models.Campaign) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a submissionBelongsToCampaignTx) Clear() error {
	return a.tx.Clear()
}

func (a submissionBelongsToCampaignTx) Count() int64 {
	return a.tx.Count()
}

func (a submissionBelongsToCampaignTx) Unscoped() *submissionBelongsToCampaignTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type submissionBelongsToRound struct {
	db *gorm.DB

	field.RelationField
}

func (a submissionBelongsToRound) Where(conds ...field.Expr) *submissionBelongsToRound {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a submissionBelongsToRound) WithContext(ctx context.Context) *submissionBelongsToRound {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a submissionBelongsToRound) Session(session *gorm.Session) *submissionBelongsToRound {
	a.db = a.db.Session(session)
	return &a
}

func (a submissionBelongsToRound) Model(m *models.Submission) *submissionBelongsToRoundTx {
	return &submissionBelongsToRoundTx{a.db.Model(m).Association(a.Name())}
}

func (a submissionBelongsToRound) Unscoped() *submissionBelongsToRound {
	a.db = a.db.Unscoped()
	return &a
}

type submissionBelongsToRoundTx struct{ tx *gorm.Association }

func (a submissionBelongsToRoundTx) Find() (result *models.Round, err error) {
	return result, a.tx.Find(&result)
}

func (a submissionBelongsToRoundTx) Append(values ...*models.Round) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a submissionBelongsToRoundTx) Replace(values ...*models.Round) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a submissionBelongsToRoundTx) Delete(values ...*models.Round) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a submissionBelongsToRoundTx) Clear() error {
	return a.tx.Clear()
}

func (a submissionBelongsToRoundTx) Count() int64 {
	return a.tx.Count()
}

func (a submissionBelongsToRoundTx) Unscoped() *submissionBelongsToRoundTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type submissionBelongsToDistributionTask struct {
	db *gorm.DB

	field.RelationField

	Submittor struct {
		field.RelationField
	}
	TaskData struct {
		field.RelationField
		Task struct {
			field.RelationField
		}
	}
}

func (a submissionBelongsToDistributionTask) Where(conds ...field.Expr) *submissionBelongsToDistributionTask {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a submissionBelongsToDistributionTask) WithContext(ctx context.Context) *submissionBelongsToDistributionTask {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a submissionBelongsToDistributionTask) Session(session *gorm.Session) *submissionBelongsToDistributionTask {
	a.db = a.db.Session(session)
	return &a
}

func (a submissionBelongsToDistributionTask) Model(m *models.Submission) *submissionBelongsToDistributionTaskTx {
	return &submissionBelongsToDistributionTaskTx{a.db.Model(m).Association(a.Name())}
}

func (a submissionBelongsToDistributionTask) Unscoped() *submissionBelongsToDistributionTask {
	a.db = a.db.Unscoped()
	return &a
}

type submissionBelongsToDistributionTaskTx struct{ tx *gorm.Association }

func (a submissionBelongsToDistributionTaskTx) Find() (result *models.Task, err error) {
	return result, a.tx.Find(&result)
}

func (a submissionBelongsToDistributionTaskTx) Append(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a submissionBelongsToDistributionTaskTx) Replace(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a submissionBelongsToDistributionTaskTx) Delete(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a submissionBelongsToDistributionTaskTx) Clear() error {
	return a.tx.Clear()
}

func (a submissionBelongsToDistributionTaskTx) Count() int64 {
	return a.tx.Count()
}

func (a submissionBelongsToDistributionTaskTx) Unscoped() *submissionBelongsToDistributionTaskTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type submissionBelongsToImportTask struct {
	db *gorm.DB

	field.RelationField
}

func (a submissionBelongsToImportTask) Where(conds ...field.Expr) *submissionBelongsToImportTask {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a submissionBelongsToImportTask) WithContext(ctx context.Context) *submissionBelongsToImportTask {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a submissionBelongsToImportTask) Session(session *gorm.Session) *submissionBelongsToImportTask {
	a.db = a.db.Session(session)
	return &a
}

func (a submissionBelongsToImportTask) Model(m *models.Submission) *submissionBelongsToImportTaskTx {
	return &submissionBelongsToImportTaskTx{a.db.Model(m).Association(a.Name())}
}

func (a submissionBelongsToImportTask) Unscoped() *submissionBelongsToImportTask {
	a.db = a.db.Unscoped()
	return &a
}

type submissionBelongsToImportTaskTx struct{ tx *gorm.Association }

func (a submissionBelongsToImportTaskTx) Find() (result *models.Task, err error) {
	return result, a.tx.Find(&result)
}

func (a submissionBelongsToImportTaskTx) Append(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a submissionBelongsToImportTaskTx) Replace(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a submissionBelongsToImportTaskTx) Delete(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a submissionBelongsToImportTaskTx) Clear() error {
	return a.tx.Clear()
}

func (a submissionBelongsToImportTaskTx) Count() int64 {
	return a.tx.Count()
}

func (a submissionBelongsToImportTaskTx) Unscoped() *submissionBelongsToImportTaskTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type submissionDo struct{ gen.DO }

type ISubmissionDo interface {
	gen.SubQuery
	Debug() ISubmissionDo
	WithContext(ctx context.Context) ISubmissionDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() ISubmissionDo
	WriteDB() ISubmissionDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) ISubmissionDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) ISubmissionDo
	Not(conds ...gen.Condition) ISubmissionDo
	Or(conds ...gen.Condition) ISubmissionDo
	Select(conds ...field.Expr) ISubmissionDo
	Where(conds ...gen.Condition) ISubmissionDo
	Order(conds ...field.Expr) ISubmissionDo
	Distinct(cols ...field.Expr) ISubmissionDo
	Omit(cols ...field.Expr) ISubmissionDo
	Join(table schema.Tabler, on ...field.Expr) ISubmissionDo
	LeftJoin(table schema.Tabler, on ...field.Expr) ISubmissionDo
	RightJoin(table schema.Tabler, on ...field.Expr) ISubmissionDo
	Group(cols ...field.Expr) ISubmissionDo
	Having(conds ...gen.Condition) ISubmissionDo
	Limit(limit int) ISubmissionDo
	Offset(offset int) ISubmissionDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) ISubmissionDo
	Unscoped() ISubmissionDo
	Create(values ...*models.Submission) error
	CreateInBatches(values []*models.Submission, batchSize int) error
	Save(values ...*models.Submission) error
	First() (*models.Submission, error)
	Take() (*models.Submission, error)
	Last() (*models.Submission, error)
	Find() ([]*models.Submission, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.Submission, err error)
	FindInBatches(result *[]*models.Submission, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*models.Submission) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) ISubmissionDo
	Assign(attrs ...field.AssignExpr) ISubmissionDo
	Joins(fields ...field.RelationField) ISubmissionDo
	Preload(fields ...field.RelationField) ISubmissionDo
	FirstOrInit() (*models.Submission, error)
	FirstOrCreate() (*models.Submission, error)
	FindByPage(offset int, limit int) (result []*models.Submission, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) ISubmissionDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (s submissionDo) Debug() ISubmissionDo {
	return s.withDO(s.DO.Debug())
}

func (s submissionDo) WithContext(ctx context.Context) ISubmissionDo {
	return s.withDO(s.DO.WithContext(ctx))
}

func (s submissionDo) ReadDB() ISubmissionDo {
	return s.Clauses(dbresolver.Read)
}

func (s submissionDo) WriteDB() ISubmissionDo {
	return s.Clauses(dbresolver.Write)
}

func (s submissionDo) Session(config *gorm.Session) ISubmissionDo {
	return s.withDO(s.DO.Session(config))
}

func (s submissionDo) Clauses(conds ...clause.Expression) ISubmissionDo {
	return s.withDO(s.DO.Clauses(conds...))
}

func (s submissionDo) Returning(value interface{}, columns ...string) ISubmissionDo {
	return s.withDO(s.DO.Returning(value, columns...))
}

func (s submissionDo) Not(conds ...gen.Condition) ISubmissionDo {
	return s.withDO(s.DO.Not(conds...))
}

func (s submissionDo) Or(conds ...gen.Condition) ISubmissionDo {
	return s.withDO(s.DO.Or(conds...))
}

func (s submissionDo) Select(conds ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.Select(conds...))
}

func (s submissionDo) Where(conds ...gen.Condition) ISubmissionDo {
	return s.withDO(s.DO.Where(conds...))
}

func (s submissionDo) Order(conds ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.Order(conds...))
}

func (s submissionDo) Distinct(cols ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.Distinct(cols...))
}

func (s submissionDo) Omit(cols ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.Omit(cols...))
}

func (s submissionDo) Join(table schema.Tabler, on ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.Join(table, on...))
}

func (s submissionDo) LeftJoin(table schema.Tabler, on ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.LeftJoin(table, on...))
}

func (s submissionDo) RightJoin(table schema.Tabler, on ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.RightJoin(table, on...))
}

func (s submissionDo) Group(cols ...field.Expr) ISubmissionDo {
	return s.withDO(s.DO.Group(cols...))
}

func (s submissionDo) Having(conds ...gen.Condition) ISubmissionDo {
	return s.withDO(s.DO.Having(conds...))
}

func (s submissionDo) Limit(limit int) ISubmissionDo {
	return s.withDO(s.DO.Limit(limit))
}

func (s submissionDo) Offset(offset int) ISubmissionDo {
	return s.withDO(s.DO.Offset(offset))
}

func (s submissionDo) Scopes(funcs ...func(gen.Dao) gen.Dao) ISubmissionDo {
	return s.withDO(s.DO.Scopes(funcs...))
}

func (s submissionDo) Unscoped() ISubmissionDo {
	return s.withDO(s.DO.Unscoped())
}

func (s submissionDo) Create(values ...*models.Submission) error {
	if len(values) == 0 {
		return nil
	}
	return s.DO.Create(values)
}

func (s submissionDo) CreateInBatches(values []*models.Submission, batchSize int) error {
	return s.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (s submissionDo) Save(values ...*models.Submission) error {
	if len(values) == 0 {
		return nil
	}
	return s.DO.Save(values)
}

func (s submissionDo) First() (*models.Submission, error) {
	if result, err := s.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.Submission), nil
	}
}

func (s submissionDo) Take() (*models.Submission, error) {
	if result, err := s.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.Submission), nil
	}
}

func (s submissionDo) Last() (*models.Submission, error) {
	if result, err := s.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.Submission), nil
	}
}

func (s submissionDo) Find() ([]*models.Submission, error) {
	result, err := s.DO.Find()
	return result.([]*models.Submission), err
}

func (s submissionDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.Submission, err error) {
	buf := make([]*models.Submission, 0, batchSize)
	err = s.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (s submissionDo) FindInBatches(result *[]*models.Submission, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return s.DO.FindInBatches(result, batchSize, fc)
}

func (s submissionDo) Attrs(attrs ...field.AssignExpr) ISubmissionDo {
	return s.withDO(s.DO.Attrs(attrs...))
}

func (s submissionDo) Assign(attrs ...field.AssignExpr) ISubmissionDo {
	return s.withDO(s.DO.Assign(attrs...))
}

func (s submissionDo) Joins(fields ...field.RelationField) ISubmissionDo {
	for _, _f := range fields {
		s = *s.withDO(s.DO.Joins(_f))
	}
	return &s
}

func (s submissionDo) Preload(fields ...field.RelationField) ISubmissionDo {
	for _, _f := range fields {
		s = *s.withDO(s.DO.Preload(_f))
	}
	return &s
}

func (s submissionDo) FirstOrInit() (*models.Submission, error) {
	if result, err := s.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.Submission), nil
	}
}

func (s submissionDo) FirstOrCreate() (*models.Submission, error) {
	if result, err := s.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.Submission), nil
	}
}

func (s submissionDo) FindByPage(offset int, limit int) (result []*models.Submission, count int64, err error) {
	result, err = s.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = s.Offset(-1).Limit(-1).Count()
	return
}

func (s submissionDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = s.Count()
	if err != nil {
		return
	}

	err = s.Offset(offset).Limit(limit).Scan(result)
	return
}

func (s submissionDo) Scan(result interface{}) (err error) {
	return s.DO.Scan(result)
}

func (s submissionDo) Delete(models ...*models.Submission) (result gen.ResultInfo, err error) {
	return s.DO.Delete(models)
}

func (s *submissionDo) withDO(do gen.Dao) *submissionDo {
	s.DO = *do.(*gen.DO)
	return s
}
