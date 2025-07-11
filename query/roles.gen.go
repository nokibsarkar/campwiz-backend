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

func newRole(db *gorm.DB, opts ...gen.DOOption) role {
	_role := role{}

	_role.roleDo.UseDB(db, opts...)
	_role.roleDo.UseModel(&models.Role{})

	tableName := _role.roleDo.TableName()
	_role.ALL = field.NewAsterisk(tableName)
	_role.RoleID = field.NewString(tableName, "role_id")
	_role.Type = field.NewString(tableName, "type")
	_role.UserID = field.NewString(tableName, "user_id")
	_role.ProjectID = field.NewString(tableName, "project_id")
	_role.TargetProjectID = field.NewString(tableName, "target_project_id")
	_role.CampaignID = field.NewString(tableName, "campaign_id")
	_role.RoundID = field.NewString(tableName, "round_id")
	_role.TotalAssigned = field.NewInt(tableName, "total_assigned")
	_role.TotalEvaluated = field.NewInt(tableName, "total_evaluated")
	_role.TotalScore = field.NewInt(tableName, "total_score")
	_role.Permission = field.NewField(tableName, "permission")
	_role.DeletedAt = field.NewField(tableName, "deleted_at")
	_role.Round = roleBelongsToRound{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Round", "models.Round"),
		Campaign: struct {
			field.RelationField
			CreatedBy struct {
				field.RelationField
				LeadingProject struct {
					field.RelationField
				}
			}
			Project struct {
				field.RelationField
			}
			LatestRound struct {
				field.RelationField
			}
			CampaignTags struct {
				field.RelationField
				Campaign struct {
					field.RelationField
				}
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
			Rounds struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Round.Campaign", "models.Campaign"),
			CreatedBy: struct {
				field.RelationField
				LeadingProject struct {
					field.RelationField
				}
			}{
				RelationField: field.NewRelation("Round.Campaign.CreatedBy", "models.User"),
				LeadingProject: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Round.Campaign.CreatedBy.LeadingProject", "models.Project"),
				},
			},
			Project: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Round.Campaign.Project", "models.Project"),
			},
			LatestRound: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Round.Campaign.LatestRound", "models.Round"),
			},
			CampaignTags: struct {
				field.RelationField
				Campaign struct {
					field.RelationField
				}
			}{
				RelationField: field.NewRelation("Round.Campaign.CampaignTags", "models.Tag"),
				Campaign: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Round.Campaign.CampaignTags.Campaign", "models.Campaign"),
				},
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
				RelationField: field.NewRelation("Round.Campaign.Roles", "models.Role"),
				Round: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Round.Campaign.Roles.Round", "models.Round"),
				},
				Campaign: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Round.Campaign.Roles.Campaign", "models.Campaign"),
				},
				User: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Round.Campaign.Roles.User", "models.User"),
				},
				Project: struct {
					field.RelationField
				}{
					RelationField: field.NewRelation("Round.Campaign.Roles.Project", "models.Project"),
				},
			},
			Rounds: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Round.Campaign.Rounds", "models.Round"),
			},
		},
		Creator: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Round.Creator", "models.User"),
		},
		DependsOnRound: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Round.DependsOnRound", "models.Round"),
		},
		Roles: struct {
			field.RelationField
		}{
			RelationField: field.NewRelation("Round.Roles", "models.Role"),
		},
	}

	_role.Campaign = roleBelongsToCampaign{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Campaign", "models.Campaign"),
	}

	_role.User = roleBelongsToUser{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("User", "models.User"),
	}

	_role.Project = roleBelongsToProject{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Project", "models.Project"),
	}

	_role.fillFieldMap()

	return _role
}

type role struct {
	roleDo

	ALL             field.Asterisk
	RoleID          field.String
	Type            field.String
	UserID          field.String
	ProjectID       field.String
	TargetProjectID field.String
	CampaignID      field.String
	RoundID         field.String
	TotalAssigned   field.Int
	TotalEvaluated  field.Int
	TotalScore      field.Int
	Permission      field.Field
	DeletedAt       field.Field
	Round           roleBelongsToRound

	Campaign roleBelongsToCampaign

	User roleBelongsToUser

	Project roleBelongsToProject

	fieldMap map[string]field.Expr
}

func (r role) Table(newTableName string) *role {
	r.roleDo.UseTable(newTableName)
	return r.updateTableName(newTableName)
}

func (r role) As(alias string) *role {
	r.roleDo.DO = *(r.roleDo.As(alias).(*gen.DO))
	return r.updateTableName(alias)
}

func (r *role) updateTableName(table string) *role {
	r.ALL = field.NewAsterisk(table)
	r.RoleID = field.NewString(table, "role_id")
	r.Type = field.NewString(table, "type")
	r.UserID = field.NewString(table, "user_id")
	r.ProjectID = field.NewString(table, "project_id")
	r.TargetProjectID = field.NewString(table, "target_project_id")
	r.CampaignID = field.NewString(table, "campaign_id")
	r.RoundID = field.NewString(table, "round_id")
	r.TotalAssigned = field.NewInt(table, "total_assigned")
	r.TotalEvaluated = field.NewInt(table, "total_evaluated")
	r.TotalScore = field.NewInt(table, "total_score")
	r.Permission = field.NewField(table, "permission")
	r.DeletedAt = field.NewField(table, "deleted_at")

	r.fillFieldMap()

	return r
}

func (r *role) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := r.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (r *role) fillFieldMap() {
	r.fieldMap = make(map[string]field.Expr, 16)
	r.fieldMap["role_id"] = r.RoleID
	r.fieldMap["type"] = r.Type
	r.fieldMap["user_id"] = r.UserID
	r.fieldMap["project_id"] = r.ProjectID
	r.fieldMap["target_project_id"] = r.TargetProjectID
	r.fieldMap["campaign_id"] = r.CampaignID
	r.fieldMap["round_id"] = r.RoundID
	r.fieldMap["total_assigned"] = r.TotalAssigned
	r.fieldMap["total_evaluated"] = r.TotalEvaluated
	r.fieldMap["total_score"] = r.TotalScore
	r.fieldMap["permission"] = r.Permission
	r.fieldMap["deleted_at"] = r.DeletedAt

}

func (r role) clone(db *gorm.DB) role {
	r.roleDo.ReplaceConnPool(db.Statement.ConnPool)
	r.Round.db = db.Session(&gorm.Session{Initialized: true})
	r.Round.db.Statement.ConnPool = db.Statement.ConnPool
	r.Campaign.db = db.Session(&gorm.Session{Initialized: true})
	r.Campaign.db.Statement.ConnPool = db.Statement.ConnPool
	r.User.db = db.Session(&gorm.Session{Initialized: true})
	r.User.db.Statement.ConnPool = db.Statement.ConnPool
	r.Project.db = db.Session(&gorm.Session{Initialized: true})
	r.Project.db.Statement.ConnPool = db.Statement.ConnPool
	return r
}

func (r role) replaceDB(db *gorm.DB) role {
	r.roleDo.ReplaceDB(db)
	r.Round.db = db.Session(&gorm.Session{})
	r.Campaign.db = db.Session(&gorm.Session{})
	r.User.db = db.Session(&gorm.Session{})
	r.Project.db = db.Session(&gorm.Session{})
	return r
}

type roleBelongsToRound struct {
	db *gorm.DB

	field.RelationField

	Campaign struct {
		field.RelationField
		CreatedBy struct {
			field.RelationField
			LeadingProject struct {
				field.RelationField
			}
		}
		Project struct {
			field.RelationField
		}
		LatestRound struct {
			field.RelationField
		}
		CampaignTags struct {
			field.RelationField
			Campaign struct {
				field.RelationField
			}
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
		Rounds struct {
			field.RelationField
		}
	}
	Creator struct {
		field.RelationField
	}
	DependsOnRound struct {
		field.RelationField
	}
	Roles struct {
		field.RelationField
	}
}

func (a roleBelongsToRound) Where(conds ...field.Expr) *roleBelongsToRound {
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

func (a roleBelongsToRound) WithContext(ctx context.Context) *roleBelongsToRound {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a roleBelongsToRound) Session(session *gorm.Session) *roleBelongsToRound {
	a.db = a.db.Session(session)
	return &a
}

func (a roleBelongsToRound) Model(m *models.Role) *roleBelongsToRoundTx {
	return &roleBelongsToRoundTx{a.db.Model(m).Association(a.Name())}
}

func (a roleBelongsToRound) Unscoped() *roleBelongsToRound {
	a.db = a.db.Unscoped()
	return &a
}

type roleBelongsToRoundTx struct{ tx *gorm.Association }

func (a roleBelongsToRoundTx) Find() (result *models.Round, err error) {
	return result, a.tx.Find(&result)
}

func (a roleBelongsToRoundTx) Append(values ...*models.Round) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a roleBelongsToRoundTx) Replace(values ...*models.Round) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a roleBelongsToRoundTx) Delete(values ...*models.Round) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a roleBelongsToRoundTx) Clear() error {
	return a.tx.Clear()
}

func (a roleBelongsToRoundTx) Count() int64 {
	return a.tx.Count()
}

func (a roleBelongsToRoundTx) Unscoped() *roleBelongsToRoundTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type roleBelongsToCampaign struct {
	db *gorm.DB

	field.RelationField
}

func (a roleBelongsToCampaign) Where(conds ...field.Expr) *roleBelongsToCampaign {
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

func (a roleBelongsToCampaign) WithContext(ctx context.Context) *roleBelongsToCampaign {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a roleBelongsToCampaign) Session(session *gorm.Session) *roleBelongsToCampaign {
	a.db = a.db.Session(session)
	return &a
}

func (a roleBelongsToCampaign) Model(m *models.Role) *roleBelongsToCampaignTx {
	return &roleBelongsToCampaignTx{a.db.Model(m).Association(a.Name())}
}

func (a roleBelongsToCampaign) Unscoped() *roleBelongsToCampaign {
	a.db = a.db.Unscoped()
	return &a
}

type roleBelongsToCampaignTx struct{ tx *gorm.Association }

func (a roleBelongsToCampaignTx) Find() (result *models.Campaign, err error) {
	return result, a.tx.Find(&result)
}

func (a roleBelongsToCampaignTx) Append(values ...*models.Campaign) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a roleBelongsToCampaignTx) Replace(values ...*models.Campaign) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a roleBelongsToCampaignTx) Delete(values ...*models.Campaign) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a roleBelongsToCampaignTx) Clear() error {
	return a.tx.Clear()
}

func (a roleBelongsToCampaignTx) Count() int64 {
	return a.tx.Count()
}

func (a roleBelongsToCampaignTx) Unscoped() *roleBelongsToCampaignTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type roleBelongsToUser struct {
	db *gorm.DB

	field.RelationField
}

func (a roleBelongsToUser) Where(conds ...field.Expr) *roleBelongsToUser {
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

func (a roleBelongsToUser) WithContext(ctx context.Context) *roleBelongsToUser {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a roleBelongsToUser) Session(session *gorm.Session) *roleBelongsToUser {
	a.db = a.db.Session(session)
	return &a
}

func (a roleBelongsToUser) Model(m *models.Role) *roleBelongsToUserTx {
	return &roleBelongsToUserTx{a.db.Model(m).Association(a.Name())}
}

func (a roleBelongsToUser) Unscoped() *roleBelongsToUser {
	a.db = a.db.Unscoped()
	return &a
}

type roleBelongsToUserTx struct{ tx *gorm.Association }

func (a roleBelongsToUserTx) Find() (result *models.User, err error) {
	return result, a.tx.Find(&result)
}

func (a roleBelongsToUserTx) Append(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a roleBelongsToUserTx) Replace(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a roleBelongsToUserTx) Delete(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a roleBelongsToUserTx) Clear() error {
	return a.tx.Clear()
}

func (a roleBelongsToUserTx) Count() int64 {
	return a.tx.Count()
}

func (a roleBelongsToUserTx) Unscoped() *roleBelongsToUserTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type roleBelongsToProject struct {
	db *gorm.DB

	field.RelationField
}

func (a roleBelongsToProject) Where(conds ...field.Expr) *roleBelongsToProject {
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

func (a roleBelongsToProject) WithContext(ctx context.Context) *roleBelongsToProject {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a roleBelongsToProject) Session(session *gorm.Session) *roleBelongsToProject {
	a.db = a.db.Session(session)
	return &a
}

func (a roleBelongsToProject) Model(m *models.Role) *roleBelongsToProjectTx {
	return &roleBelongsToProjectTx{a.db.Model(m).Association(a.Name())}
}

func (a roleBelongsToProject) Unscoped() *roleBelongsToProject {
	a.db = a.db.Unscoped()
	return &a
}

type roleBelongsToProjectTx struct{ tx *gorm.Association }

func (a roleBelongsToProjectTx) Find() (result *models.Project, err error) {
	return result, a.tx.Find(&result)
}

func (a roleBelongsToProjectTx) Append(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a roleBelongsToProjectTx) Replace(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a roleBelongsToProjectTx) Delete(values ...*models.Project) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a roleBelongsToProjectTx) Clear() error {
	return a.tx.Clear()
}

func (a roleBelongsToProjectTx) Count() int64 {
	return a.tx.Count()
}

func (a roleBelongsToProjectTx) Unscoped() *roleBelongsToProjectTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type roleDo struct{ gen.DO }

type IRoleDo interface {
	gen.SubQuery
	Debug() IRoleDo
	WithContext(ctx context.Context) IRoleDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IRoleDo
	WriteDB() IRoleDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IRoleDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IRoleDo
	Not(conds ...gen.Condition) IRoleDo
	Or(conds ...gen.Condition) IRoleDo
	Select(conds ...field.Expr) IRoleDo
	Where(conds ...gen.Condition) IRoleDo
	Order(conds ...field.Expr) IRoleDo
	Distinct(cols ...field.Expr) IRoleDo
	Omit(cols ...field.Expr) IRoleDo
	Join(table schema.Tabler, on ...field.Expr) IRoleDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IRoleDo
	RightJoin(table schema.Tabler, on ...field.Expr) IRoleDo
	Group(cols ...field.Expr) IRoleDo
	Having(conds ...gen.Condition) IRoleDo
	Limit(limit int) IRoleDo
	Offset(offset int) IRoleDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IRoleDo
	Unscoped() IRoleDo
	Create(values ...*models.Role) error
	CreateInBatches(values []*models.Role, batchSize int) error
	Save(values ...*models.Role) error
	First() (*models.Role, error)
	Take() (*models.Role, error)
	Last() (*models.Role, error)
	Find() ([]*models.Role, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.Role, err error)
	FindInBatches(result *[]*models.Role, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*models.Role) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IRoleDo
	Assign(attrs ...field.AssignExpr) IRoleDo
	Joins(fields ...field.RelationField) IRoleDo
	Preload(fields ...field.RelationField) IRoleDo
	FirstOrInit() (*models.Role, error)
	FirstOrCreate() (*models.Role, error)
	FindByPage(offset int, limit int) (result []*models.Role, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IRoleDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (r roleDo) Debug() IRoleDo {
	return r.withDO(r.DO.Debug())
}

func (r roleDo) WithContext(ctx context.Context) IRoleDo {
	return r.withDO(r.DO.WithContext(ctx))
}

func (r roleDo) ReadDB() IRoleDo {
	return r.Clauses(dbresolver.Read)
}

func (r roleDo) WriteDB() IRoleDo {
	return r.Clauses(dbresolver.Write)
}

func (r roleDo) Session(config *gorm.Session) IRoleDo {
	return r.withDO(r.DO.Session(config))
}

func (r roleDo) Clauses(conds ...clause.Expression) IRoleDo {
	return r.withDO(r.DO.Clauses(conds...))
}

func (r roleDo) Returning(value interface{}, columns ...string) IRoleDo {
	return r.withDO(r.DO.Returning(value, columns...))
}

func (r roleDo) Not(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Not(conds...))
}

func (r roleDo) Or(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Or(conds...))
}

func (r roleDo) Select(conds ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Select(conds...))
}

func (r roleDo) Where(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Where(conds...))
}

func (r roleDo) Order(conds ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Order(conds...))
}

func (r roleDo) Distinct(cols ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Distinct(cols...))
}

func (r roleDo) Omit(cols ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Omit(cols...))
}

func (r roleDo) Join(table schema.Tabler, on ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Join(table, on...))
}

func (r roleDo) LeftJoin(table schema.Tabler, on ...field.Expr) IRoleDo {
	return r.withDO(r.DO.LeftJoin(table, on...))
}

func (r roleDo) RightJoin(table schema.Tabler, on ...field.Expr) IRoleDo {
	return r.withDO(r.DO.RightJoin(table, on...))
}

func (r roleDo) Group(cols ...field.Expr) IRoleDo {
	return r.withDO(r.DO.Group(cols...))
}

func (r roleDo) Having(conds ...gen.Condition) IRoleDo {
	return r.withDO(r.DO.Having(conds...))
}

func (r roleDo) Limit(limit int) IRoleDo {
	return r.withDO(r.DO.Limit(limit))
}

func (r roleDo) Offset(offset int) IRoleDo {
	return r.withDO(r.DO.Offset(offset))
}

func (r roleDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IRoleDo {
	return r.withDO(r.DO.Scopes(funcs...))
}

func (r roleDo) Unscoped() IRoleDo {
	return r.withDO(r.DO.Unscoped())
}

func (r roleDo) Create(values ...*models.Role) error {
	if len(values) == 0 {
		return nil
	}
	return r.DO.Create(values)
}

func (r roleDo) CreateInBatches(values []*models.Role, batchSize int) error {
	return r.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (r roleDo) Save(values ...*models.Role) error {
	if len(values) == 0 {
		return nil
	}
	return r.DO.Save(values)
}

func (r roleDo) First() (*models.Role, error) {
	if result, err := r.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.Role), nil
	}
}

func (r roleDo) Take() (*models.Role, error) {
	if result, err := r.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.Role), nil
	}
}

func (r roleDo) Last() (*models.Role, error) {
	if result, err := r.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.Role), nil
	}
}

func (r roleDo) Find() ([]*models.Role, error) {
	result, err := r.DO.Find()
	return result.([]*models.Role), err
}

func (r roleDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.Role, err error) {
	buf := make([]*models.Role, 0, batchSize)
	err = r.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (r roleDo) FindInBatches(result *[]*models.Role, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return r.DO.FindInBatches(result, batchSize, fc)
}

func (r roleDo) Attrs(attrs ...field.AssignExpr) IRoleDo {
	return r.withDO(r.DO.Attrs(attrs...))
}

func (r roleDo) Assign(attrs ...field.AssignExpr) IRoleDo {
	return r.withDO(r.DO.Assign(attrs...))
}

func (r roleDo) Joins(fields ...field.RelationField) IRoleDo {
	for _, _f := range fields {
		r = *r.withDO(r.DO.Joins(_f))
	}
	return &r
}

func (r roleDo) Preload(fields ...field.RelationField) IRoleDo {
	for _, _f := range fields {
		r = *r.withDO(r.DO.Preload(_f))
	}
	return &r
}

func (r roleDo) FirstOrInit() (*models.Role, error) {
	if result, err := r.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.Role), nil
	}
}

func (r roleDo) FirstOrCreate() (*models.Role, error) {
	if result, err := r.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.Role), nil
	}
}

func (r roleDo) FindByPage(offset int, limit int) (result []*models.Role, count int64, err error) {
	result, err = r.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = r.Offset(-1).Limit(-1).Count()
	return
}

func (r roleDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = r.Count()
	if err != nil {
		return
	}

	err = r.Offset(offset).Limit(limit).Scan(result)
	return
}

func (r roleDo) Scan(result interface{}) (err error) {
	return r.DO.Scan(result)
}

func (r roleDo) Delete(models ...*models.Role) (result gen.ResultInfo, err error) {
	return r.DO.Delete(models)
}

func (r *roleDo) withDO(do gen.Dao) *roleDo {
	r.DO = *do.(*gen.DO)
	return r
}
