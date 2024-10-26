package register

import (
	"embed"
	"errors"
	"sync"

	vh "github.com/vera-byte/vgo-lib/pkg/vgohttp"
)

// APIRouter api路由者
type APIRouter interface {
	Route(r *vh.VGoHttp)
}

// var apiRoutes = make([]APIRouter, 0)

// // Add 添加api
// func Add(r APIRouter) {
// 	apiRoutes = append(apiRoutes, r)
// }

// var taskRoutes = make([]TaskRouter, 0)

// // GetRoutes 获取所有路由者
// func GetRoutes() []APIRouter {
// 	return apiRoutes
// }

// // TaskRouter task路由者
// type TaskRouter interface {
// 	RegisterTasks()
// }

// // AddTask 添加任务
// func AddTask(task TaskRouter) {
// 	taskRoutes = append(taskRoutes, task)
// }

// // GetTasks 获取所有任务
// func GetTasks() []TaskRouter {
// 	return taskRoutes
// }

type ModuleFnc func(ctx interface{}) Module

var modules = make([]ModuleFnc, 0)

type IMDatasourceType int

const (
	IMDatasourceTypeNone        IMDatasourceType = iota
	IMDatasourceTypeSubscribers                  = 1
	IMDatasourceTypeChannelInfo                  = 1 << 1
	IMDatasourceTypeBlacklist                    = 1 << 2
	IMDatasourceTypeWhitelist                    = 1 << 3
	IMDatasourceTypeSystemUIDs                   = 1 << 4
)

func (i IMDatasourceType) Has(d IMDatasourceType) bool {
	return i&d == d
}

var (
	ErrDatasourceNotProcess error = errors.New("datasource not process")
)

// 模块
type Module struct {
	// 模块名称
	Name string
	// api 路由
	SetupAPI func() APIRouter
	// task 路由
	SetupTask func() TaskRouter
	// 服务
	// sql目录
	SQLDir *SQLFS
	// swagger文件
	Swagger string
	// 服务
	Service interface{}
	// 事件
	Start func() error
	Stop  func() error
}

func AddModule(moduleFnc func(ctx interface{}) Module) {
	modules = append(modules, moduleFnc)
}

type SQLFS struct {
	embed.FS
}

func NewSQLFS(fs embed.FS) *SQLFS {

	return &SQLFS{
		FS: fs,
	}
}

var once sync.Once
var moduleList []Module

func GetModules(ctx any) []Module {

	once.Do(func() {
		for _, m := range modules {
			moduleList = append(moduleList, m(ctx))
		}
	})

	return moduleList
}

func GetModuleByName(name string, ctx any) Module {

	for _, m := range moduleList {
		if m.Name == name {
			return m
		}
	}
	return Module{}
}

// GetService 获取服务
func GetService(name string) interface{} {
	for _, m := range moduleList {
		if m.Name == name {
			return m.Service
		}
	}
	return nil
}

// TaskRouter task路由者
type TaskRouter interface {
	RegisterTasks()
}
