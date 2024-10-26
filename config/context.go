package config

import (
	"sync"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/bwmarrin/snowflake"
	"github.com/gocraft/dbr/v2"
	"github.com/olivere/elastic"
	"github.com/opentracing/opentracing-go"
	"github.com/vera-byte/vgo-lib/common"
	"github.com/vera-byte/vgo-lib/pkg/cache"
	"github.com/vera-byte/vgo-lib/pkg/db"
	"github.com/vera-byte/vgo-lib/pkg/log"
	"github.com/vera-byte/vgo-lib/pkg/pool"
	"github.com/vera-byte/vgo-lib/pkg/redis"
	"github.com/vera-byte/vgo-lib/pkg/wkhttp"
)

// Context 配置上下文
type Context struct {
	cfg          *Config
	mySQLSession *dbr.Session
	redisCache   *common.RedisCache
	memoryCache  cache.Cache
	log.Log
	EventPool     pool.Collector
	elasticClient *elastic.Client
	UserIDGen     *snowflake.Node          // 消息ID生成器
	tracer        *Tracer                  // 调用链追踪
	aysncTask     *AsyncTask               // 异步任务
	timingWheel   *timingwheel.TimingWheel // Time wheel delay task

	httpRouter *wkhttp.WKHttp

	valueMap  sync.Map
	SetupTask bool // 是否安装task
}

// NewContext NewContext
func NewContext(cfg *Config) *Context {
	userIDGen, err := snowflake.NewNode(int64(cfg.Cluster.NodeID))
	if err != nil {
		panic(err)
	}
	c := &Context{
		cfg:         cfg,
		UserIDGen:   userIDGen,
		Log:         log.NewTLog("Context"),
		EventPool:   pool.StartDispatcher(cfg.EventPoolSize),
		aysncTask:   NewAsyncTask(cfg),
		timingWheel: timingwheel.NewTimingWheel(cfg.TimingWheelTick.Duration, cfg.TimingWheelSize),
		valueMap:    sync.Map{},
	}
	c.tracer, err = NewTracer(cfg)
	if err != nil {
		panic(err)
	}
	opentracing.SetGlobalTracer(c.tracer)
	c.timingWheel.Start()
	return c
}

// GetConfig 获取配置信息
func (c *Context) GetConfig() *Config {
	return c.cfg
}

// NewMySQL 创建mysql数据库实例
func (c *Context) NewMySQL() *dbr.Session {

	if c.mySQLSession == nil {
		c.mySQLSession = db.NewMySQL(c.cfg.DB.MySQLAddr, c.cfg.DB.MySQLMaxOpenConns, c.cfg.DB.MySQLMaxIdleConns, c.cfg.DB.MySQLConnMaxLifetime)
	}

	return c.mySQLSession
}

// AsyncTask 异步任务
func (c *Context) AsyncTask() *AsyncTask {
	return c.aysncTask
}

// Tracer Tracer
func (c *Context) Tracer() *Tracer {
	return c.tracer
}

// DB DB
func (c *Context) DB() *dbr.Session {
	return c.NewMySQL()
}

// NewRedisCache 创建一个redis缓存
func (c *Context) NewRedisCache() *common.RedisCache {
	if c.redisCache == nil {
		c.redisCache = common.NewRedisCache(c.cfg.DB.RedisAddr, c.cfg.DB.RedisPass)
	}
	return c.redisCache
}

// NewMemoryCache 创建一个内存缓存
func (c *Context) NewMemoryCache() cache.Cache {
	if c.memoryCache == nil {
		c.memoryCache = common.NewMemoryCache()
	}
	return c.memoryCache
}

// Cache 缓存
func (c *Context) Cache() cache.Cache {
	return c.NewRedisCache()
}

// 认证中间件
func (c *Context) AuthMiddleware(r *wkhttp.WKHttp) wkhttp.HandlerFunc {

	return r.AuthMiddleware(c.Cache(), c.cfg.Cache.TokenCachePrefix)
}

// GetRedisConn GetRedisConn
func (c *Context) GetRedisConn() *redis.Conn {
	return c.NewRedisCache().GetRedisConn()
}

// Schedule 延迟任务
func (c *Context) Schedule(interval time.Duration, f func()) *timingwheel.Timer {
	return c.timingWheel.ScheduleFunc(&everyScheduler{
		Interval: interval,
	}, f)
}

func (c *Context) GetHttpRoute() *wkhttp.WKHttp {
	return c.httpRouter
}

func (c *Context) SetHttpRoute(r *wkhttp.WKHttp) {
	c.httpRouter = r
}

func (c *Context) SetValue(value interface{}, key string) {
	c.valueMap.Store(key, value)
}

func (c *Context) Value(key string) any {
	v, _ := c.valueMap.Load(key)
	return v
}

// EventCommit 事件提交
type EventCommit func(err error)

// EventListener EventListener
type EventListener func(data []byte, commit EventCommit)

var eventListeners = map[string][]EventListener{}

// AddEventListener  添加事件监听
func (c *Context) AddEventListener(event string, listener EventListener) {
	listeners := eventListeners[event]
	if listeners == nil {
		listeners = make([]EventListener, 0)
	}
	listeners = append(listeners, listener)
	eventListeners[event] = listeners
}

// GetEventListeners 获取某个事件
func (c *Context) GetEventListeners(event string) []EventListener {
	return eventListeners[event]
}

type everyScheduler struct {
	Interval time.Duration
}

func (s *everyScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}
