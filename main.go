package main

import (
	"fmt"
	"os"
	"strconv"
	"taskboard-api-go/controller/api"
	"taskboard-api-go/controller/boards"
	"taskboard-api-go/controller/tasks"
	"taskboard-api-go/controller/users"
	"taskboard-api-go/model"
	"taskboard-api-go/orm"
	"taskboard-api-go/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Init database
	fmt.Println("Initializing database...")
	databasePath := "./taskboard.sqlite3"
	err := orm.Init(databasePath)
	if err != nil {
		fmt.Printf("Failed to initialize database. error:%+v\n", err)
		return
	}

	// Create or Update tables
	err = orm.Migrate(
		&model.User{},
		&model.Task{},
		&model.Board{},
	)
	if err != nil {
		fmt.Printf("Failed to update tables. error:%+v\n", err)
		return
	}

	// Create system boards(Icebox, Todo, Doing, Done)
	tx := orm.GetDB().Begin()
	srvc := service.NewBoardService(tx)
	if err = srvc.CreateSystemBoards(); err != nil {
		fmt.Printf("Failed to create system boards. error:%+v\n", err)
		api.Rollback(tx)
		return
	}
	if err = api.Commit(tx); err != nil {
		return
	}

	// Init router of REST apis
	router := gin.Default()
	//config := cors.DefaultConfig()
	//config.AllowAllOrigins = true
	//config.AllowHeaders = []string{"Content-Type"}
	//router.Use(cors.New(config))
	router.Use(cors.Default())
	// Include static/avatars
	router.Static("/taskboard/static", "./static")

	// Register api path
	routeGroup := router.Group("/taskboard")
	users.EndPoint.RegisterRoute(routeGroup)
	boards.EndPoint.RegisterRoute(routeGroup)
	tasks.EndPoint.RegisterRoute(routeGroup)

	// Set listening host:port
	url := getListeningURL()
	// Start server
	fmt.Printf("Taskboard api server is starting... listening %s\n", url)
	err = router.Run(url)
	if err != nil {
		fmt.Printf("Failed to start api server. error:%+v\n", err)
		return
	}
}

func getListeningURL() string {
	host := os.Getenv("TASKBOARD_API_SERVER_HOST")
	if host == "" {
		fmt.Println("Environment variable [TASKBOARD_API_SERVER_HOST] is not set or invalid. localhost is used as default")
	}
	portEnv := os.Getenv("TASKBOARD_API_SERVER_PORT")
	port := 0
	if portEnv != "" {
		port, err := strconv.Atoi(portEnv)
		if err != nil || port < 0 {
			port = 0
		}
	}
	if port == 0 {
		fmt.Println("Environment variable [TASKBOARD_API_SERVER_PORT] is not set or invalid, 7000 port is used as default.")
		port = 7000
	}
	return fmt.Sprintf("%s:%d", host, port)
}
