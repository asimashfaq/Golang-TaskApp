package main

import (
	"fmt"
	loger "internal/log"
	taskmodel "internal/model/task"
	taskservice "internal/service/task"
	"internal/shared"
	"os"
	"strconv"
	"time"

	"github.com/ftloc/exception"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var session *shared.Session
var tasks_service *taskservice.TaskService
var count = 0

func start() {
	context, cancel, c := shared.CreateContext(5 * time.Second)
	var err error
	go func() {
		defer func() {
			if err := recover(); err != nil {
				loger.Error("Got Panic Exception:", err)
				context.Err()
				cancel()

			}
		}()
		session, err = shared.NewSessesion(c, context, "mongodb://tasksuser:tabletask@159.65.37.36")
		if err == nil {
			cancel()
			tasks_service = taskservice.NewTasKService(session, "tasksapp", "tasks")
			go func() {
				time.Sleep(10 * time.Second)
				for {
					go checkorReconnect()
					time.Sleep(20 * time.Second)
				}

			}()

		}

	}()
	select {
	case <-context.Done():

		if context.Err().Error() != "context canceled" {
			loger.Error("Context Failed to connect to db, Timeout Exceeds, Check your Internet connection or Connection Url")
			os.Exit(1)
		}

	case <-c:
		fmt.Println("Connection Open success! Task Done")

	}

}
func checkorReconnect() {
	count++
	loger.Info("checkORReconnect called", count)
	context, cancel, c := shared.CreateContext(10 * time.Second)

	exception.Try(func() {

		go func() {
			defer func() {
				if err := recover(); err != nil {
					loger.Error("Got Panic Exception:", err)
					cancel()
				}
			}()
			fmt.Println(session.Session.Ping())
			if session.Session.Ping() != nil {
				loger.Warn("Ping failed now will reconnect")
				context2, cancel2, c2 := shared.CreateContext(5 * time.Second)

				go func() {
					defer func() {
						if err := recover(); err != nil {
							loger.Error("Got Panic Exception:", err)
							cancel2()
							cancel()

						}
					}()
					session, _ = shared.NewSessesion(c2, context2, "mongodb://tasksuser:tabletask@159.65.37.36")
					tasks_service = taskservice.NewTasKService(session, "tasksapp", "tasks")
					cancel2()
					cancel()

				}()
				shared.WatchContextForDBConnection(context2, c2)

			} else {
				cancel()

			}

		}()
		shared.WatchContextForDBConnection(context, c)

	}).CatchAll(func(e interface{}) {
		//	cancel()
		context.Err()
		loger.Warn("Inside exception handelr", e.(error))

	}).Go()
	loger.Info("Context1 is finished asked to reconnect")

}
func main() {
	start()
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.POST("/task", CreateTask)
	e.GET("/tasks", GetAllTasks)
	e.DELETE("/task/:id", DeleteTask)
	e.PUT("/task", TaskProgressUpdate)
	e.Logger.Fatal(e.Start(":8080"))

}

func CreateTask(c echo.Context) error {
	payload := new(shared.TaskPayload)
	c.Bind(&payload)
	var rep map[string]interface{}
	exception.Try(func() {
		if len(payload.Days) > 0 {
			days := []taskmodel.TaskProgressFrequency{}
			for _, day := range payload.Days {
				days = append(days, taskmodel.TaskProgressFrequency{
					Day:    day,
					Status: false,
				})
			}
			response, err := tasks_service.InsertNewTask(taskmodel.TaskModel{
				Title:     payload.Title,
				Frequency: days,
				Color:     payload.Color,
			})
			if err == nil {
				rep = gin.H{
					"success": true,
					"result":  response,
				}
			}
		} else {
			rep = gin.H{
				"success": true,
				"result":  "No days are provided",
			}
		}

	}).CatchAll(func(e interface{}) {
		loger.Error(e.(error))
		rep = gin.H{
			"success": false,
			"result":  e.(error).Error(),
		}
	}).Go()
	return c.JSON(200, rep)
}
func GetAllTasks(c echo.Context) error {
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	var rep map[string]interface{}
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"result":  "Invalid Limit",
		})
	}
	skip, err := strconv.Atoi(c.QueryParam("skip"))
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"result":  "Invalid Limit",
		})
	}

	exception.Try(func() {
		tasks, count := tasks_service.GetAllTask(limit, skip)
		if len(tasks) == 0 {
			rep = gin.H{
				"success": true,
				"result":  "No record found",
			}
		} else {
			rep = gin.H{
				"success": true,
				"count":   count,
				"result":  tasks,
			}
		}
	}).CatchAll(func(e interface{}) {
		loger.Error(e.(error))
		rep = gin.H{
			"success": false,
			"result":  e.(error).Error(),
		}
	}).Go()

	return c.JSON(200, rep)
}
func DeleteTask(c echo.Context) error {
	id := c.Param("id")
	var rep map[string]interface{}
	exception.Try(func() {
		response, _ := tasks_service.DeleteTaskById(id)
		if response == true {
			rep = gin.H{
				"success": true,
				"result":  "Task Deleted Successfully",
			}
		}
		rep = gin.H{
			"success": false,
			"result":  "Unable to delete check ur ObjectID or Record does not exists",
		}

	}).CatchAll(func(e interface{}) {
		loger.Error(e.(error))
		rep = gin.H{
			"success": false,
			"result":  e.(error).Error(),
		}
	}).Go()
	return c.JSON(200, rep)
}
func TaskProgressUpdate(c echo.Context) error {
	payload := new(shared.TaskProgress)
	c.Bind(payload)
	statuts := tasks_service.UpdateTaskProgress(*payload)
	return c.JSON(200, gin.H{
		"success": statuts,
	})

}
