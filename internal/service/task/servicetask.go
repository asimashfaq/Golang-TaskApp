package taskservice

import (
	"fmt"
	taskmodel "internal/model/task"
	"internal/shared"
	"strings"
	"time"

	loger "internal/log"

	"github.com/ftloc/exception"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TaskService struct {
	collection *mgo.Collection
}

func NewTasKService(session *shared.Session, dbName string, collectionName string) *TaskService {
	collection := session.GetCollection(dbName, collectionName)
	collection.EnsureIndex(taskmodel.TaskModelIndex())
	return &TaskService{collection}
}

func (t *TaskService) InsertNewTask(data taskmodel.TaskModel) (taskmodel.TaskModel, error) {

	data.Alias = strings.ToLower(strings.Replace(data.Title, " ", "-", -1))
	isExist, err := t.GetTaskByAlias(data.Alias)
	if err != nil {
		loger.Error(err)
		exception.Throw(fmt.Errorf("Something goes worng %+v", errors.WithStack(err)))
	}
	if isExist.Alias == data.Alias {
		exception.Throw(fmt.Errorf("Task Already Exists Or Something goes worng"))
	}

	context, cancel, c := shared.CreateContext(5 * time.Second)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				loger.Error("Got Panic Exception:", err)
				cancel()

			}
		}()
		err = t.collection.Insert(&data)
		if err != nil {
			exception.Throw(fmt.Errorf("Failed to Insert %+v", errors.WithStack(err)))
		}
		cancel()

	}()
	shared.WatchContextForDBConnection(context, c)
	data, err = t.GetTaskByAlias(data.Alias)
	if err != nil {
		exception.Throw(fmt.Errorf("Failed to Read after Insert %+v", errors.WithStack(err)))
	}
	return data, nil
}
func (t *TaskService) GetTaskByAlias(taskName string) (taskmodel.TaskModel, error) {
	model := taskmodel.TaskModel{}
	query := bson.M{
		"alias": taskName,
	}
	var err error

	context, cancel, c := shared.CreateContext(5 * time.Second)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				loger.Error("Got Panic Exception:", err)
				cancel()

			}
		}()
		err = t.collection.Find(query).One(&model)
		if err != nil {
			if err.Error() != "not found" {
				cancel()
				exception.Throw(fmt.Errorf("Failed to Query %+v", errors.WithStack(err)))
			} else {
				err = nil
			}

		}
		cancel()

	}()
	shared.WatchContextForDBConnection(context, c)
	return model, err
}
func (t *TaskService) GetAllTask(limit int, skip int) ([]taskmodel.TaskModel, int) {
	loger.Info("Trying to get All tasks")
	if limit < 10 {
		exception.Throw(fmt.Errorf("Invalid Limit, Limit Must be Less than 10"))
	}
	tasks := []taskmodel.TaskModel{}
	context, cancel, c := shared.CreateContext(5 * time.Second)
	var count int
	go func() {
		defer func() {
			if err := recover(); err != nil {
				loger.Error("Got Panic Exception:", err)
				cancel()

			}
		}()
		if err := t.collection.Find(nil).Sort("-_id").Limit(limit).Skip(skip).All(&tasks); err != nil {
			loger.Error("Error while getting tasks", err)
			exception.Throw(fmt.Errorf("Failed to Query %+v", errors.WithStack(err)))

		}
		count, _ = t.collection.Find(nil).Count()

		cancel()

	}()
	shared.WatchContextForDBConnection(context, c)
	return tasks, count
}
func (t *TaskService) DeleteTaskById(id string) (bool, error) {
	isvalidId := bson.IsObjectIdHex(id)
	if !isvalidId {
		return false, nil
	}

	context, cancel, c := shared.CreateContext(5 * time.Second)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				loger.Error("Got Panic Exception:", err)
				cancel()

			}
		}()
		err := t.collection.RemoveId(bson.ObjectIdHex(id))
		if err != nil {
			exception.Throw(fmt.Errorf("Failed to Query or Record %+v", errors.WithStack(err)))

		}
		cancel()

	}()
	shared.WatchContextForDBConnection(context, c)

	return true, nil
}
func (t *TaskService) UpdateTaskProgress(taskprogress shared.TaskProgress) bool {

	context, cancel, c := shared.CreateContext(5 * time.Second)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				loger.Error("Got Panic Exception:", err)
				cancel()

			}
		}()
		query := bson.M{
			"_id":     bson.ObjectId(taskprogress.ID),
			"tpf.day": taskprogress.Day,
		}
		update := bson.M{
			"$set": bson.M{
				"tpf.$.status": taskprogress.Status,
			},
		}
		err := t.collection.Update(query, update)
		if err != nil {
			loger.Error("Failed to update task", err)
			exception.Throw(fmt.Errorf("Failed to update task %+v", errors.WithStack(err)))
		}
		cancel()

	}()
	shared.WatchContextForDBConnection(context, c)

	return true
}
