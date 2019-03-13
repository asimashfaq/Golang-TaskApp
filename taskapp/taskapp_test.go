package main

import (
	"fmt"
	loger "internal/log"
	taskmodel "internal/model/task"
	taskservice "internal/service/task"
	"internal/shared"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ftloc/exception"
)

func InitalizeService() *taskservice.TaskService {
	var session *shared.Session
	context, cancel, c := shared.CreateContext(5 * time.Second)
	go func() {
		session, _ = shared.NewSessesion(c, context, "mongodb://tasksuser:tabletask@159.65.37.36")

		cancel()

	}()
	shared.WatchContextForDBConnection(context, c)
	return taskservice.NewTasKService(session, "tasksapp", "tasks")
}

/*func TestCreateTask(t *testing.T) {
	tasks_service := InitalizeService()
	result, er := tasks_service.InsertNewTask(taskmodel.TaskModel{

		Title: "Test Task1",
		Frequency: []taskmodel.TaskProgressFrequency{
			taskmodel.TaskProgressFrequency{Day: "Mon", Status: false},
			taskmodel.TaskProgressFrequency{Day: "Wed", Status: false},
			taskmodel.TaskProgressFrequency{Day: "Fri", Status: false},
		},
		Color: "#000000",
	})
	fmt.Println(result, er)
}*/
func TestCreateTaskAlreadyExist(t *testing.T) {

	exception.Try(func() {
		tasks_service := InitalizeService()
		result, er := tasks_service.InsertNewTask(taskmodel.TaskModel{

			Title: "Test Task1",
			Frequency: []taskmodel.TaskProgressFrequency{
				taskmodel.TaskProgressFrequency{Day: "Mon", Status: false},
				taskmodel.TaskProgressFrequency{Day: "Wed", Status: false},
				taskmodel.TaskProgressFrequency{Day: "Fri", Status: false},
			},
			Color: "#000000",
		})
		fmt.Println(result, er)
	}).CatchAll(func(e interface{}) {
		loger.Error(e)
		assert.Equal(t, e.(error).Error(), "Task Already Exists Or Something goes worng")
	}).Go()

}
func TestDeleteTask(t *testing.T) {
	tasks_service := InitalizeService()
	result, _ := tasks_service.DeleteTaskById("5c85eb359d527b271c820842")
	fmt.Println(result)
}
func TestGetALlTask(t *testing.T) {
	tasks_service := InitalizeService()
	result, _ := tasks_service.GetAllTask(10, 0)
	fmt.Println(result)

}
