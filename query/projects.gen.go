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

func newProject(db *gorm.DB, opts ...gen.DOOption) project {
	_project := project{}

	_project.projectDo.UseDB(db, opts...)
	_project.projectDo.UseModel(&models.Project{})

	tableName := _project.projectDo.TableName()
	_project.ALL = field.NewAsterisk(tableName)
	_project.ProjectID = field.NewString(tableName, "project_id")
	_project.Name = field.NewString(tableName, "name")
	_project.LogoURL = field.NewString(tableName, "logo_url")
	_project.Link = field.NewString(tableName, "link")
	_project.CreatedByID = field.NewString(tableName, "created_by_id")
	_project.CreatedAt = field.NewTime(tableName, "created_at")

	_project.fillFieldMap()

	return _project
}

type project struct {
	projectDo

	ALL         field.Asterisk
	ProjectID   field.String
	Name        field.String
	LogoURL     field.String
	Link        field.String
	CreatedByID field.String
	CreatedAt   field.Time

	fieldMap map[string]field.Expr
}

func (p project) Table(newTableName string) *project {
	p.projectDo.UseTable(newTableName)
	return p.updateTableName(newTableName)
}

func (p project) As(alias string) *project {
	p.projectDo.DO = *(p.projectDo.As(alias).(*gen.DO))
	return p.updateTableName(alias)
}

func (p *project) updateTableName(table string) *project {
	p.ALL = field.NewAsterisk(table)
	p.ProjectID = field.NewString(table, "project_id")
	p.Name = field.NewString(table, "name")
	p.LogoURL = field.NewString(table, "logo_url")
	p.Link = field.NewString(table, "link")
	p.CreatedByID = field.NewString(table, "created_by_id")
	p.CreatedAt = field.NewTime(table, "created_at")

	p.fillFieldMap()

	return p
}

func (p *project) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := p.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (p *project) fillFieldMap() {
	p.fieldMap = make(map[string]field.Expr, 6)
	p.fieldMap["project_id"] = p.ProjectID
	p.fieldMap["name"] = p.Name
	p.fieldMap["logo_url"] = p.LogoURL
	p.fieldMap["link"] = p.Link
	p.fieldMap["created_by_id"] = p.CreatedByID
	p.fieldMap["created_at"] = p.CreatedAt
}

func (p project) clone(db *gorm.DB) project {
	p.projectDo.ReplaceConnPool(db.Statement.ConnPool)
	return p
}

func (p project) replaceDB(db *gorm.DB) project {
	p.projectDo.ReplaceDB(db)
	return p
}

type projectDo struct{ gen.DO }

type IProjectDo interface {
	gen.SubQuery
	Debug() IProjectDo
	WithContext(ctx context.Context) IProjectDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IProjectDo
	WriteDB() IProjectDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IProjectDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IProjectDo
	Not(conds ...gen.Condition) IProjectDo
	Or(conds ...gen.Condition) IProjectDo
	Select(conds ...field.Expr) IProjectDo
	Where(conds ...gen.Condition) IProjectDo
	Order(conds ...field.Expr) IProjectDo
	Distinct(cols ...field.Expr) IProjectDo
	Omit(cols ...field.Expr) IProjectDo
	Join(table schema.Tabler, on ...field.Expr) IProjectDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IProjectDo
	RightJoin(table schema.Tabler, on ...field.Expr) IProjectDo
	Group(cols ...field.Expr) IProjectDo
	Having(conds ...gen.Condition) IProjectDo
	Limit(limit int) IProjectDo
	Offset(offset int) IProjectDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IProjectDo
	Unscoped() IProjectDo
	Create(values ...*models.Project) error
	CreateInBatches(values []*models.Project, batchSize int) error
	Save(values ...*models.Project) error
	First() (*models.Project, error)
	Take() (*models.Project, error)
	Last() (*models.Project, error)
	Find() ([]*models.Project, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.Project, err error)
	FindInBatches(result *[]*models.Project, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*models.Project) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IProjectDo
	Assign(attrs ...field.AssignExpr) IProjectDo
	Joins(fields ...field.RelationField) IProjectDo
	Preload(fields ...field.RelationField) IProjectDo
	FirstOrInit() (*models.Project, error)
	FirstOrCreate() (*models.Project, error)
	FindByPage(offset int, limit int) (result []*models.Project, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IProjectDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (p projectDo) Debug() IProjectDo {
	return p.withDO(p.DO.Debug())
}

func (p projectDo) WithContext(ctx context.Context) IProjectDo {
	return p.withDO(p.DO.WithContext(ctx))
}

func (p projectDo) ReadDB() IProjectDo {
	return p.Clauses(dbresolver.Read)
}

func (p projectDo) WriteDB() IProjectDo {
	return p.Clauses(dbresolver.Write)
}

func (p projectDo) Session(config *gorm.Session) IProjectDo {
	return p.withDO(p.DO.Session(config))
}

func (p projectDo) Clauses(conds ...clause.Expression) IProjectDo {
	return p.withDO(p.DO.Clauses(conds...))
}

func (p projectDo) Returning(value interface{}, columns ...string) IProjectDo {
	return p.withDO(p.DO.Returning(value, columns...))
}

func (p projectDo) Not(conds ...gen.Condition) IProjectDo {
	return p.withDO(p.DO.Not(conds...))
}

func (p projectDo) Or(conds ...gen.Condition) IProjectDo {
	return p.withDO(p.DO.Or(conds...))
}

func (p projectDo) Select(conds ...field.Expr) IProjectDo {
	return p.withDO(p.DO.Select(conds...))
}

func (p projectDo) Where(conds ...gen.Condition) IProjectDo {
	return p.withDO(p.DO.Where(conds...))
}

func (p projectDo) Order(conds ...field.Expr) IProjectDo {
	return p.withDO(p.DO.Order(conds...))
}

func (p projectDo) Distinct(cols ...field.Expr) IProjectDo {
	return p.withDO(p.DO.Distinct(cols...))
}

func (p projectDo) Omit(cols ...field.Expr) IProjectDo {
	return p.withDO(p.DO.Omit(cols...))
}

func (p projectDo) Join(table schema.Tabler, on ...field.Expr) IProjectDo {
	return p.withDO(p.DO.Join(table, on...))
}

func (p projectDo) LeftJoin(table schema.Tabler, on ...field.Expr) IProjectDo {
	return p.withDO(p.DO.LeftJoin(table, on...))
}

func (p projectDo) RightJoin(table schema.Tabler, on ...field.Expr) IProjectDo {
	return p.withDO(p.DO.RightJoin(table, on...))
}

func (p projectDo) Group(cols ...field.Expr) IProjectDo {
	return p.withDO(p.DO.Group(cols...))
}

func (p projectDo) Having(conds ...gen.Condition) IProjectDo {
	return p.withDO(p.DO.Having(conds...))
}

func (p projectDo) Limit(limit int) IProjectDo {
	return p.withDO(p.DO.Limit(limit))
}

func (p projectDo) Offset(offset int) IProjectDo {
	return p.withDO(p.DO.Offset(offset))
}

func (p projectDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IProjectDo {
	return p.withDO(p.DO.Scopes(funcs...))
}

func (p projectDo) Unscoped() IProjectDo {
	return p.withDO(p.DO.Unscoped())
}

func (p projectDo) Create(values ...*models.Project) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Create(values)
}

func (p projectDo) CreateInBatches(values []*models.Project, batchSize int) error {
	return p.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (p projectDo) Save(values ...*models.Project) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Save(values)
}

func (p projectDo) First() (*models.Project, error) {
	if result, err := p.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.Project), nil
	}
}

func (p projectDo) Take() (*models.Project, error) {
	if result, err := p.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.Project), nil
	}
}

func (p projectDo) Last() (*models.Project, error) {
	if result, err := p.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.Project), nil
	}
}

func (p projectDo) Find() ([]*models.Project, error) {
	result, err := p.DO.Find()
	return result.([]*models.Project), err
}

func (p projectDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.Project, err error) {
	buf := make([]*models.Project, 0, batchSize)
	err = p.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (p projectDo) FindInBatches(result *[]*models.Project, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return p.DO.FindInBatches(result, batchSize, fc)
}

func (p projectDo) Attrs(attrs ...field.AssignExpr) IProjectDo {
	return p.withDO(p.DO.Attrs(attrs...))
}

func (p projectDo) Assign(attrs ...field.AssignExpr) IProjectDo {
	return p.withDO(p.DO.Assign(attrs...))
}

func (p projectDo) Joins(fields ...field.RelationField) IProjectDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Joins(_f))
	}
	return &p
}

func (p projectDo) Preload(fields ...field.RelationField) IProjectDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Preload(_f))
	}
	return &p
}

func (p projectDo) FirstOrInit() (*models.Project, error) {
	if result, err := p.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.Project), nil
	}
}

func (p projectDo) FirstOrCreate() (*models.Project, error) {
	if result, err := p.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.Project), nil
	}
}

func (p projectDo) FindByPage(offset int, limit int) (result []*models.Project, count int64, err error) {
	result, err = p.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = p.Offset(-1).Limit(-1).Count()
	return
}

func (p projectDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = p.Count()
	if err != nil {
		return
	}

	err = p.Offset(offset).Limit(limit).Scan(result)
	return
}

func (p projectDo) Scan(result interface{}) (err error) {
	return p.DO.Scan(result)
}

func (p projectDo) Delete(models ...*models.Project) (result gen.ResultInfo, err error) {
	return p.DO.Delete(models)
}

func (p *projectDo) withDO(do gen.Dao) *projectDo {
	p.DO = *do.(*gen.DO)
	return p
}
