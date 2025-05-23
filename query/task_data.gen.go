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

func newTaskData(db *gorm.DB, opts ...gen.DOOption) taskData {
	_taskData := taskData{}

	_taskData.taskDataDo.UseDB(db, opts...)
	_taskData.taskDataDo.UseModel(&models.TaskData{})

	tableName := _taskData.taskDataDo.TableName()
	_taskData.ALL = field.NewAsterisk(tableName)
	_taskData.DataID = field.NewString(tableName, "data_id")
	_taskData.TaskID = field.NewString(tableName, "task_id")
	_taskData.Key = field.NewString(tableName, "key")
	_taskData.Value = field.NewString(tableName, "value")
	_taskData.IsOutput = field.NewBool(tableName, "is_output")
	_taskData.Task = taskDataBelongsToTask{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Task", "models.Task"),
		Submittor: struct {
			field.RelationField
			LeadingProject struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Task.Submittor", "models.User"),
			LeadingProject: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Task.Submittor.LeadingProject", "models.Project"),
			},
		},
		TaskData: struct {
			field.RelationField
			Task struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Task.TaskData", "models.TaskData"),
			Task: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Task.TaskData.Task", "models.Task"),
			},
		},
	}

	_taskData.fillFieldMap()

	return _taskData
}

type taskData struct {
	taskDataDo

	ALL      field.Asterisk
	DataID   field.String
	TaskID   field.String
	Key      field.String
	Value    field.String
	IsOutput field.Bool
	Task     taskDataBelongsToTask

	fieldMap map[string]field.Expr
}

func (t taskData) Table(newTableName string) *taskData {
	t.taskDataDo.UseTable(newTableName)
	return t.updateTableName(newTableName)
}

func (t taskData) As(alias string) *taskData {
	t.taskDataDo.DO = *(t.taskDataDo.As(alias).(*gen.DO))
	return t.updateTableName(alias)
}

func (t *taskData) updateTableName(table string) *taskData {
	t.ALL = field.NewAsterisk(table)
	t.DataID = field.NewString(table, "data_id")
	t.TaskID = field.NewString(table, "task_id")
	t.Key = field.NewString(table, "key")
	t.Value = field.NewString(table, "value")
	t.IsOutput = field.NewBool(table, "is_output")

	t.fillFieldMap()

	return t
}

func (t *taskData) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := t.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (t *taskData) fillFieldMap() {
	t.fieldMap = make(map[string]field.Expr, 6)
	t.fieldMap["data_id"] = t.DataID
	t.fieldMap["task_id"] = t.TaskID
	t.fieldMap["key"] = t.Key
	t.fieldMap["value"] = t.Value
	t.fieldMap["is_output"] = t.IsOutput

}

func (t taskData) clone(db *gorm.DB) taskData {
	t.taskDataDo.ReplaceConnPool(db.Statement.ConnPool)
	t.Task.db = db.Session(&gorm.Session{Initialized: true})
	t.Task.db.Statement.ConnPool = db.Statement.ConnPool
	return t
}

func (t taskData) replaceDB(db *gorm.DB) taskData {
	t.taskDataDo.ReplaceDB(db)
	t.Task.db = db.Session(&gorm.Session{})
	return t
}

type taskDataBelongsToTask struct {
	db *gorm.DB

	field.RelationField

	Submittor struct {
		field.RelationField
		LeadingProject struct {
			field.RelationField
		}
	}
	TaskData struct {
		field.RelationField
		Task struct {
			field.RelationField
		}
	}
}

func (a taskDataBelongsToTask) Where(conds ...field.Expr) *taskDataBelongsToTask {
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

func (a taskDataBelongsToTask) WithContext(ctx context.Context) *taskDataBelongsToTask {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a taskDataBelongsToTask) Session(session *gorm.Session) *taskDataBelongsToTask {
	a.db = a.db.Session(session)
	return &a
}

func (a taskDataBelongsToTask) Model(m *models.TaskData) *taskDataBelongsToTaskTx {
	return &taskDataBelongsToTaskTx{a.db.Model(m).Association(a.Name())}
}

func (a taskDataBelongsToTask) Unscoped() *taskDataBelongsToTask {
	a.db = a.db.Unscoped()
	return &a
}

type taskDataBelongsToTaskTx struct{ tx *gorm.Association }

func (a taskDataBelongsToTaskTx) Find() (result *models.Task, err error) {
	return result, a.tx.Find(&result)
}

func (a taskDataBelongsToTaskTx) Append(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a taskDataBelongsToTaskTx) Replace(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a taskDataBelongsToTaskTx) Delete(values ...*models.Task) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a taskDataBelongsToTaskTx) Clear() error {
	return a.tx.Clear()
}

func (a taskDataBelongsToTaskTx) Count() int64 {
	return a.tx.Count()
}

func (a taskDataBelongsToTaskTx) Unscoped() *taskDataBelongsToTaskTx {
	a.tx = a.tx.Unscoped()
	return &a
}

type taskDataDo struct{ gen.DO }

type ITaskDataDo interface {
	gen.SubQuery
	Debug() ITaskDataDo
	WithContext(ctx context.Context) ITaskDataDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() ITaskDataDo
	WriteDB() ITaskDataDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) ITaskDataDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) ITaskDataDo
	Not(conds ...gen.Condition) ITaskDataDo
	Or(conds ...gen.Condition) ITaskDataDo
	Select(conds ...field.Expr) ITaskDataDo
	Where(conds ...gen.Condition) ITaskDataDo
	Order(conds ...field.Expr) ITaskDataDo
	Distinct(cols ...field.Expr) ITaskDataDo
	Omit(cols ...field.Expr) ITaskDataDo
	Join(table schema.Tabler, on ...field.Expr) ITaskDataDo
	LeftJoin(table schema.Tabler, on ...field.Expr) ITaskDataDo
	RightJoin(table schema.Tabler, on ...field.Expr) ITaskDataDo
	Group(cols ...field.Expr) ITaskDataDo
	Having(conds ...gen.Condition) ITaskDataDo
	Limit(limit int) ITaskDataDo
	Offset(offset int) ITaskDataDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) ITaskDataDo
	Unscoped() ITaskDataDo
	Create(values ...*models.TaskData) error
	CreateInBatches(values []*models.TaskData, batchSize int) error
	Save(values ...*models.TaskData) error
	First() (*models.TaskData, error)
	Take() (*models.TaskData, error)
	Last() (*models.TaskData, error)
	Find() ([]*models.TaskData, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.TaskData, err error)
	FindInBatches(result *[]*models.TaskData, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*models.TaskData) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) ITaskDataDo
	Assign(attrs ...field.AssignExpr) ITaskDataDo
	Joins(fields ...field.RelationField) ITaskDataDo
	Preload(fields ...field.RelationField) ITaskDataDo
	FirstOrInit() (*models.TaskData, error)
	FirstOrCreate() (*models.TaskData, error)
	FindByPage(offset int, limit int) (result []*models.TaskData, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) ITaskDataDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (t taskDataDo) Debug() ITaskDataDo {
	return t.withDO(t.DO.Debug())
}

func (t taskDataDo) WithContext(ctx context.Context) ITaskDataDo {
	return t.withDO(t.DO.WithContext(ctx))
}

func (t taskDataDo) ReadDB() ITaskDataDo {
	return t.Clauses(dbresolver.Read)
}

func (t taskDataDo) WriteDB() ITaskDataDo {
	return t.Clauses(dbresolver.Write)
}

func (t taskDataDo) Session(config *gorm.Session) ITaskDataDo {
	return t.withDO(t.DO.Session(config))
}

func (t taskDataDo) Clauses(conds ...clause.Expression) ITaskDataDo {
	return t.withDO(t.DO.Clauses(conds...))
}

func (t taskDataDo) Returning(value interface{}, columns ...string) ITaskDataDo {
	return t.withDO(t.DO.Returning(value, columns...))
}

func (t taskDataDo) Not(conds ...gen.Condition) ITaskDataDo {
	return t.withDO(t.DO.Not(conds...))
}

func (t taskDataDo) Or(conds ...gen.Condition) ITaskDataDo {
	return t.withDO(t.DO.Or(conds...))
}

func (t taskDataDo) Select(conds ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.Select(conds...))
}

func (t taskDataDo) Where(conds ...gen.Condition) ITaskDataDo {
	return t.withDO(t.DO.Where(conds...))
}

func (t taskDataDo) Order(conds ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.Order(conds...))
}

func (t taskDataDo) Distinct(cols ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.Distinct(cols...))
}

func (t taskDataDo) Omit(cols ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.Omit(cols...))
}

func (t taskDataDo) Join(table schema.Tabler, on ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.Join(table, on...))
}

func (t taskDataDo) LeftJoin(table schema.Tabler, on ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.LeftJoin(table, on...))
}

func (t taskDataDo) RightJoin(table schema.Tabler, on ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.RightJoin(table, on...))
}

func (t taskDataDo) Group(cols ...field.Expr) ITaskDataDo {
	return t.withDO(t.DO.Group(cols...))
}

func (t taskDataDo) Having(conds ...gen.Condition) ITaskDataDo {
	return t.withDO(t.DO.Having(conds...))
}

func (t taskDataDo) Limit(limit int) ITaskDataDo {
	return t.withDO(t.DO.Limit(limit))
}

func (t taskDataDo) Offset(offset int) ITaskDataDo {
	return t.withDO(t.DO.Offset(offset))
}

func (t taskDataDo) Scopes(funcs ...func(gen.Dao) gen.Dao) ITaskDataDo {
	return t.withDO(t.DO.Scopes(funcs...))
}

func (t taskDataDo) Unscoped() ITaskDataDo {
	return t.withDO(t.DO.Unscoped())
}

func (t taskDataDo) Create(values ...*models.TaskData) error {
	if len(values) == 0 {
		return nil
	}
	return t.DO.Create(values)
}

func (t taskDataDo) CreateInBatches(values []*models.TaskData, batchSize int) error {
	return t.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (t taskDataDo) Save(values ...*models.TaskData) error {
	if len(values) == 0 {
		return nil
	}
	return t.DO.Save(values)
}

func (t taskDataDo) First() (*models.TaskData, error) {
	if result, err := t.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.TaskData), nil
	}
}

func (t taskDataDo) Take() (*models.TaskData, error) {
	if result, err := t.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.TaskData), nil
	}
}

func (t taskDataDo) Last() (*models.TaskData, error) {
	if result, err := t.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.TaskData), nil
	}
}

func (t taskDataDo) Find() ([]*models.TaskData, error) {
	result, err := t.DO.Find()
	return result.([]*models.TaskData), err
}

func (t taskDataDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.TaskData, err error) {
	buf := make([]*models.TaskData, 0, batchSize)
	err = t.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (t taskDataDo) FindInBatches(result *[]*models.TaskData, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return t.DO.FindInBatches(result, batchSize, fc)
}

func (t taskDataDo) Attrs(attrs ...field.AssignExpr) ITaskDataDo {
	return t.withDO(t.DO.Attrs(attrs...))
}

func (t taskDataDo) Assign(attrs ...field.AssignExpr) ITaskDataDo {
	return t.withDO(t.DO.Assign(attrs...))
}

func (t taskDataDo) Joins(fields ...field.RelationField) ITaskDataDo {
	for _, _f := range fields {
		t = *t.withDO(t.DO.Joins(_f))
	}
	return &t
}

func (t taskDataDo) Preload(fields ...field.RelationField) ITaskDataDo {
	for _, _f := range fields {
		t = *t.withDO(t.DO.Preload(_f))
	}
	return &t
}

func (t taskDataDo) FirstOrInit() (*models.TaskData, error) {
	if result, err := t.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.TaskData), nil
	}
}

func (t taskDataDo) FirstOrCreate() (*models.TaskData, error) {
	if result, err := t.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.TaskData), nil
	}
}

func (t taskDataDo) FindByPage(offset int, limit int) (result []*models.TaskData, count int64, err error) {
	result, err = t.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = t.Offset(-1).Limit(-1).Count()
	return
}

func (t taskDataDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = t.Count()
	if err != nil {
		return
	}

	err = t.Offset(offset).Limit(limit).Scan(result)
	return
}

func (t taskDataDo) Scan(result interface{}) (err error) {
	return t.DO.Scan(result)
}

func (t taskDataDo) Delete(models ...*models.TaskData) (result gen.ResultInfo, err error) {
	return t.DO.Delete(models)
}

func (t *taskDataDo) withDO(do gen.Dao) *taskDataDo {
	t.DO = *do.(*gen.DO)
	return t
}
