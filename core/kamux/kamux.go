package kamux

import (
	"flag"
	"net/http"
	"os"
	"regexp"
	"sync"

	"github.com/kamalshkeir/kago/core/orm"
	"github.com/kamalshkeir/kago/core/settings"
	"github.com/kamalshkeir/kago/core/shell"
	"github.com/kamalshkeir/kago/core/utils"
	"github.com/kamalshkeir/kago/core/utils/logger"
	"golang.org/x/net/websocket"
)

const (
	GET int = iota
	POST
	PUT
	PATCH
	DELETE
	WS
	SSE
)
var methods = map[int]string{
	GET:"GET",
	POST:"POST",
	PUT:"PUT",
	PATCH:"PATCH",
	DELETE:"DELETE",
	WS:"WS",
	SSE:"SSE",
}
// Handler
type Handler func(c *Context)
type WsHandler func(c *WsContext)

// Router
type Router struct {
	Routes       map[int][]Route
	DefaultRoute Handler
	Server *http.Server
}

// Route
type Route struct {
	Method  string
	Pattern *regexp.Regexp
	Handler 
	WsHandler
	Clients map[string]*websocket.Conn
	AllowedOrigines []string
}

// New Create New Router from env file default: '.env'
func New(envFiles ...string) *Router {
	var wg sync.WaitGroup
	app := &Router{
		Routes: map[int][]Route{},
		DefaultRoute: func(c *Context) {
			c.Status(404).Text("Page Not Found")
		},
	}
	
	if len(envFiles) > 0 {
		app.LoadEnv(envFiles...)
	} else {
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			if os.Getenv("DB_NAME") == "" && os.Getenv("DB_DSN") == "" {
				logger.Warn("Environment variables not loaded, you can copy it from generated assets folder and rename it to .env, or set them manualy")
			}
		} else {
			app.LoadEnv(".env")
		}
	}

	// check flags
	wg.Add(1)
	go func() {
		h := flag.String("h","localhost","overwrite host")
		p := flag.String("p","9313","overwrite port number")
		logs := flag.Bool("logs",false,"overwrite settings.Config.Logs for router /logs")
		monitoring := flag.Bool("monitoring",false,"set settings.Config.Monitoring for prometheus and grafana /metrics")
		docs := flag.Bool("docs",false,"set settings.Config.Docs for prometheus and grafana /docs")
		profiler := flag.Bool("profiler",false,"set settings.Config.Profiler for pprof  /debug/pprof")
		flag.Parse()
		
		settings.Config.Logs=*logs
		settings.Config.Monitoring=*monitoring
		settings.Config.Docs=*docs
		settings.Config.Profiler=*profiler
		if *p != "9313" {
			settings.Config.Port=*p
		}
		if *h != "localhost" && *h != "127.0.0.1" && *h != "" {
			settings.Config.Host=*h
		} else {
			settings.Config.Host="localhost"
		}
		wg.Done()
	}()
	
	// load translations
	wg.Add(1)
	go func() {
		LoadTranslations()
		wg.Done()
	}()
	wg.Wait()

	err := orm.InitDB()
	if err != nil {
		if settings.Config.Db.Name == "" && settings.Config.Db.DSN == "" {
			logger.Warn("Environment variables not loaded, you can copy it from generated assets folder and rename it to .env, or set them manualy")
		} else {
			logger.Error(err)
		}
	} 
	// migrate initial models
	err = orm.Migrate()
	logger.CheckError(err)
	// init orm shell
	if shell.InitShell() {os.Exit(0)}
	utils.PrintServerStart()
	return app
}

// handle a route
func (router *Router) handle(method int,pattern string, handler Handler,wshandler WsHandler,allowed []string) {
	re := regexp.MustCompile(adaptParams(pattern))
	route := Route{Method: methods[method],Pattern: re, Handler: handler, WsHandler: wshandler,Clients: nil,AllowedOrigines: []string{}}
	if len(allowed) > 0 && method != GET {
		route.AllowedOrigines = append(route.AllowedOrigines, allowed...)
	}
	if method == WS {
		route.Clients=map[string]*websocket.Conn{}
	}
	if _,ok := router.Routes[method];!ok {
		router.Routes[method]=[]Route{}
	}
	if len(router.Routes[method]) == 0 {
		router.Routes[method] = append(router.Routes[method], route)
		return
	}
	for i,rt := range router.Routes[method] {
		if rt.Pattern.String() == re.String() {
			router.Routes[method]=append(router.Routes[method][:i],router.Routes[method][i+1:]...)
		} 
	}
	router.Routes[method] = append(router.Routes[method], route)
}

// GET handle GET to a route
func (router *Router) GET(pattern string, handler Handler) {
	router.handle(GET,pattern,handler,nil,nil)
}

// POST handle POST to a route
func (router *Router) POST(pattern string, handler Handler, allowed_origines ...string) {
	router.handle(POST,pattern,handler,nil,allowed_origines)
}

// PUT handle PUT to a route
func (router *Router) PUT(pattern string, handler Handler, allowed_origines ...string) {
	router.handle(PUT,pattern,handler,nil,allowed_origines)
}

// PATCH handle PATCH to a route
func (router *Router) PATCH(pattern string, handler Handler, allowed_origines ...string) {
	router.handle(PATCH,pattern,handler,nil,allowed_origines)
}

// DELETE handle DELETE to a route
func (router *Router) DELETE(pattern string, handler Handler, allowed_origines ...string) {
	router.handle(DELETE,pattern,handler,nil,allowed_origines)
}

// WS handle WS connection on a pattern
func (router *Router) WS(pattern string, wsHandler WsHandler, allowed_origines ...string) {
	router.handle(WS,pattern,nil,wsHandler,allowed_origines)
}

// SSE handle SSE to a route
func (router *Router) SSE(pattern string, handler Handler, allowed_origines ...string) {
	router.handle(SSE,pattern,handler,nil,allowed_origines)
}


