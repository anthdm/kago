<div style="display:flex;justify-content-center;">
	<img src="./logo.png" width="auto" style="margin:0 auto 0 auto;"/>
</div>
<br>

<div style="display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
	<img src="https://img.shields.io/github/go-mod/go-version/kamalshkeir/kago" width="auto" height="20px">
	<img src="https://img.shields.io/github/languages/code-size/kamalshkeir/kago" width="auto" height="20px">
	<img src="https://img.shields.io/badge/License-BSD%20v3-blue.svg" width="auto" height="20px">
	<img src="https://img.shields.io/badge/License-BSD%20v3-blue.svg" width="auto" height="20px">
	<img src="https://img.shields.io/github/v/tag/kamalshkeir/kago" width="auto" height="20px">
	<img src="https://img.shields.io/github/stars/kamalshkeir/kago?style=social" width="auto" height="20px">
	<img src="https://img.shields.io/github/forks/kamalshkeir/kago?style=social" width="auto" height="20px">
</div>
<br>
<div style="display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
	<img src="https://img.shields.io/badge/my_portfolio-000?style=for-the-badge&logo=ko-fi&logoColor=white" href="https://kamalshkeir.github.io/" width="auto" height="auto">
	<img src="https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge&logo=linkedin&logoColor=white" href="https://www.linkedin.com/in/kamal-shkeir/" width="auto" height="auto">
</div>

# KaGo Web Framework 
<br>
KaGo is a high-level web framework, that encourages clean and rapid development.

You can literally get up and running using two lines of code, easier than Django and with a compiled language performance.

Kago offer you :
- Fully editable CRUD Admin Dashboard (assets folder)
- Ready to use Progressive Web App Support (pure js, without workbox)
- Realtime Logs at `/logs` running with flag `go run main.go --logs`
- Convenient router that handle params with regex type checking, Websockets and SSE protocols also 
- Interactive Shell `go run main.go shell`
- Performant and Easy ORM and SQL builder using go generics (working with sqlite, mysql and postgres)
- OpenApi docs at `/docs` running with flag `go run main.go --docs` with dedicated package docs to help you manipulate docs.json file
- AES encryption for authentication sessions using package encryptor
- Argon2 hashing using package hash
- EnvLoader using package envloader
- Internal Eventbus using go routines and channels (very powerful), use cases are endless
- Monitoring Prometheus/grafana `/metrics` running with flag `go run main.go --monitoring`
- Profiler golang official debug pprof `/debug/pprof/(heap|profile\trace)` running with flag `go run main.go --profiler`

If you aim performance and good productivity, you will love KaGo.

Many features will be added in the future, like:
- the possibility to choose a theme and create your own
- automatic backups and send by email
- Mongo DB support (soon)

Join our  [discussions here](https://github.com/kamalshkeir/kago/discussions/1)


## Installation

To install it, you need to Go installed and set your Go workspace first.

1. You first need [Go](https://golang.org/) installed (**version 1.18+ is required**), then you can use the below Go command to install kago.

```sh
$ go get github.com/kamalshkeir/kago
```

2. Import it in your code:

```go
import "github.com/kamalshkeir/kago"
```

## Quick start

make sure you have git installed

Create main.go:
```go
package main

import "github.com/kamalshkeir/kago"

func main() {
	app := kago.New()
	app.Run()
}
```

#### 1- running 'go run main.go' the first time, will clone assets folder with all static and template files for admin using git
```shell
go run main.go
```

#### 2- make sure you copy 'assets/.env.example' beside your main at the root folder and rename it to '.env'

#### 3- go to the shell to create admin account
```shell
go run main.go shell    

-> createsuperuser
```


#### 4- you can change port and host by putting Env Vars 'HOST' and 'PORT' or using flags:
```zsh
go run main.go -h 0.0.0.0 -p 3333
```
###### will run on http://[privateIP]:3333 


## Generated Admin Dashboard
##### you can easily override any handler of any url by creating a new one with the same method and path.
### for example, to override the handler at GET /admin/login:
```go
app.GET("/admin/login",func(c *kamux.Context) {
	...
})
// this will add another handler, and remove the old one completely 
// so it is very safe to do it this way also
```

### The other way: i make all handlers and all middlewares as variables to give you full control on the behavior of any 
### handler/middleware already written by me
```go
// all these handlers can be overriden
r.GET("/mon/ping",func(c *kamux.Context) {c.Status(200).Text("pong")})
r.GET("/offline",OfflineView) 
r.GET("/manifest.webmanifest",ManifestView) 
r.GET("/sw.js",ServiceWorkerView) 
r.GET("/robots.txt",RobotsTxtView) 
r.GET("/admin", middlewares.Admin(IndexView))
r.GET("/admin/login",middlewares.Auth(LoginView))
r.POST("/admin/login",middlewares.Auth(LoginPOSTView))
r.GET("/admin/logout", LogoutView)
r.POST("/admin/delete/row", middlewares.Admin(DeleteRowPost))
r.POST("/admin/update/row", middlewares.Admin(UpdateRowPost))
r.POST("/admin/create/row", middlewares.Admin(CreateModelView))
r.POST("/admin/drop/table", middlewares.Admin(DropTablePost))
r.GET("/admin/table/model:str", middlewares.Admin(AllModelsGet))
r.POST("/admin/table/model:str", middlewares.Admin(AllModelsPost))
r.GET("/admin/get/model:str/id:int", middlewares.Admin(SingleModelGet))
r.GET("/admin/export/table:str", middlewares.Admin(ExportView))
r.POST("/admin/import", middlewares.Admin(ImportView))
r.GET("/logs",middlewares.Admin(LogsGetView))
r.SSE("/sse/logs",middlewares.Admin(LogsSSEView))

// Example : how to override a handler
admin.LoginView=func(c *kamux.Context) {
	...		
}

// Example : how to override a middleware
middlewares.Auth = func(handler kamux.Handler) kamux.Handler { // handlerFunc
		...
}
// Example : how to override a Global middleware
middlewares.GZIP = func(handler http.Handler) http.Handler { // Handler
		...
}
```

## Routing
### Using GET, POST, PUT, PATCH, DELETE

```go
package main

import (
	"github.com/kamalshkeir/kago"
	"github.com/kamalshkeir/kago/core/kamux"
	"github.com/kamalshkeir/kago/core/middlewares"
)


func main() {
    // Creates a KaGo router
	app := kago.New()

    // You can add GLOBAL middlewares easily  (GZIP,CORS,CSRF,LIMITER,RECOVERY)
	app.UseMiddlewares(middlewares.GZIP)

	// OR middleware for single handler (Auth,Admin,BasicAuth)
	// Auth ensure user is authenticated and pass to c.Html '.user' and '.request', so accessible in all templates
	app.GET("/",middlewares.Auth(IndexHandler))

	app.POST("/somePost", posting)
	app.PUT("/somePut", putting)
	app.DELETE("/someDelete", deleting)
	app.PATCH("/somePatch", patching)
	
	app.Run()
}

var IndexHandler = func(c *kamux.Context) {
    if param1,ok := c.Params["param1"];ok {
        c.Status(200).Json(map[string]any{
            "param1":param1,
        }) // send json
    } else {
		// P.S: 
        c.Status(404).Text("Not Found") // Status will not write status to header,only when chained with Json or Text will be executed
		c.WriteHeader(404) // will set the status header
    }
}
```

## Websockets + Server Sent Events
### After this one, you will stop being afraid to try websockets everywhere

```go
func main() {
	app := kago.New()
	
	// no need to upgrade the request , all you need to worry about is
	// inside this handler, you can enjoy realtime communication
	app.WS("/ws/test",func(c *kamux.WsContext) {
		rand := utils.GenerateRandomString(5)
		c.AddClient(rand) // add connection to broadcast list

		// listen for messages coming from 1 user
		for {
			// receive Json
			mapStringAny,err := c.ReceiveJson()
			// receive Text
			str,err := c.ReceiveText()
			if err != nil {
				// on error you can remove client from broadcastList and break the loop
				c.RemoveRequester(rand)
				break
			}

			// send Json to current user
			err = c.Json(map[string]any{
				"Hello":"World",
			})

			// send Text to current user
			err = c.Text("any data string")

			// broadcast to all connected users
			c.Broadcast(map[string]any{
				"you can send":"struct insetead of maps here",
			})

			// broadcast to all connected users except current user, the one who send the last message
			c.BroadcastExceptCaller(map[string]any{
				"you can send":"struct insetead of maps here",
			})

		}
	})
	
	app.Run()
}
```
### Server Sent Events
```go
func main() {
	app := kago.New()
	
	// will be hitted every 1-2 sec, you can check anything if change and send data on the fly using c.StreamResponse
	app.SSE("/sse/logs",func(c *kamux.Context) {
		lenStream := len(logger.StreamLogs)
		if lenStream > 0 {
			lastOne := lenStream-1
			err := c.StreamResponse(logger.StreamLogs[lastOne])
			if err == nil{
				logger.StreamLogs=[]string{}
			}
		}
	})
	
	app.Run()
}
```

## Parameters (path + query)

```go
func main() {
	app := kago.New()

    // Query string parameters are parsed using the existing underlying request object
    // request url : /?page=3
    app.GET("/",func(c *kamux.Context) {
		page := c.QueryParam("page")
		if page != "" {
			c.Status(200).Json(map[string]any{
				"page":page,
			})
		} else {
			c.Status(404)
		}
	})

    // This handler will match /anyString but will not match /
    // accepted param Type: string,slug,int,float and validated on the go, before it hit the handler
    app.POST("/param1:slug",func(c *kamux.Context) {
		if param1,ok := c.Params["param1"];ok {
			c.Json(map[string]any{
				"param1":param1,
			})
		} else {
			c.Status(404).Text("Not Found")
		}
	})

	// OR 

	// param1 can be ascii, no symbole
	app.PATCH("/test/:param1",func(c *kamux.Context) {
		if param1,ok := c.Params["param1"];ok {
			c.Json(map[string]any{
				"param1":param1,
			})
		} else {
			c.Status(404).Text("Not Found")
		}
	})

	// param1 can be ascii, no symbole -> /test/anything will work
	app.POST("/test/param1:str",func(c *kamux.Context) 

	//  /test/anything will not work, should be int, validated with regex -> /100 will work
	app.POST("/test/param1:int",func(c *kamux.Context) 

	//  2 digit float -> /test/3.45 will work
	app.POST("/test/param1:float",func(c *kamux.Context)

	// slug param should match abcd-efgh, no spacing 
	app.POST("/test/param1:slug",func(c *kamux.Context) 


	app.Run()
}
```


### Router Http Context
###### There is also WsContext seen above

```go
func main() {
	app := kago.New()

    // Query string parameters are parsed using the existing underlying request object
    // request url : /?page=3
    app.GET("/",func(c *kamux.Context) {
		page := c.QueryParam("page")
		if page != "" {
			c.Status(200).Json(map[string]any{
				"page":page,
			})
		} else {
			c.writeHeader(404)
		}
	})

    // This handler will match /anyString but will not match /
    // accepted param Type: string,slug,int,float and validated on the go, before it hit the handler
    app.POST("/param1:slug",func(c *kamux.Context) {
		if param1,ok := c.Params["param1"];ok {
			c.Json(map[string]any{
				"param1":param1,
			})
		} else {
			c.Status(404).Text("Not Found")
		}
	})

	// and many more
	c.Status(200).JsonIndent(code int, body any)
	c.Status(200).Html(template_name string, data map[string]any)
	c.Status(301).Redirect(path string) // redirect to path
	c.BodyJson() map[string]any // get request body
	c.StreamResponse(response string) error //SSE
	c.ServeFile("application/json; charset=utf-8", "./test.json")
	c.ServeEmbededFile(content_type string,embed_file []byte)
	c.UploadFile(received_filename,folder_out string, acceptedFormats ...string) (string,[]byte,error) // UploadFile upload received_filename into folder_out and return url,fileByte,error
	c.UploadFiles(received_filenames []string,folder_out string, acceptedFormats ...string) ([]string,[][]byte,error) // UploadFilse handle also if it's the same name but multiple files or multiple names multiple files
	c.DeleteFile(path string) error
	c.Download(data_bytes []byte, asFilename string)
	c.EnableTranslations() // EnableTranslations get user ip, then location country using nmap, so don't use it if u don't have it install, and then it parse csv file to find the language spoken in this country, to finaly set cookie 'lang' to 'en' or 'fr'... 
	c.GetUserIP() string // get user ip

	app.Run()
}
```

### Multipart/Urlencoded Form

```go
func main() {
	app := kago.New()

	app.POST("/ajax", func(c *kamux.Context) {
        // get json body to map[string]any
		requestData := c.BodyJson()
        if email,ok := requestData["email"]; ok {
            ...
        }

        if err != nil {
            c.Status(400).Json(map[string]any{
			    "error":"User doesn not Exist",
		    })
        }
	})
	app.Run()
}
```

### Upload file
```go
func main() {
	app := kago.New()

	// kamux.MultipartSize = 10<<20 memory limit set to 10Mb
	app.POST("/upload", func(c *kamux.Context) {
		// UploadFileFromFormData upload received_filename into folder_out and return url,fileByte,error
        //c.UploadFile(received_filename,folder_out string, acceptedFormats ...string) (string,[]byte,error)
        pathToFile,dataBytes,err := c.UploadFile("filename_from_form","images","png","jpg")
		// you can save pathToFile in db from here
		c.Status(200).Text(file.Filename + " uploaded")
	})


	app.Run(":8080")
}
```

### Cookies
```go
func main() {
	app := kago.New()

	kamux.COOKIE_EXPIRE= time.Now().Add(7 * 24 * time.Hour)
	
	app.POST("/ajax", func(c *kamux.Context) {
        // get json body to map[string]any
		requestData := c.BodyJson()
        if email,ok := requestData["email"]; ok {
            ...
        }

		// set cookie
		c.SetCookie(key,value string)
		c.GetCookie(key string)
		c.DeleteCookie(key string)
	})

	app.Run(":8080")
}
```

### HTML functions maps
```go

// To add one, automatically loaded into your html
app.NewFuncMap(funcName string, function any)

/* FUNC MAPS */
var functions = template.FuncMap{
	"isBool": func(something any) bool {
	"isTrue": func(something any) bool {
	"contains": func(str string, substrings ...string) bool {
	"startWith": func(str string, substrings ...string) bool {
	"finishWith": func(str string, substrings ...string) bool {
	"generateUUID": func() template.HTML {
	"add": func(a int,b int) int {
	"safe": func(str string) template.HTML {
	"timeFormat":func (t any) string {
	"truncate": func(str any,size int) any {
	"csrf_token":func (r *http.Request) template.HTML {
	"translateFromRequest":func (translation string, request *http.Request) any {
	"translateFromLang":func (translation,language  string) any {
}

```

## Add Custom Static And Templates Folders
##### you can build all your static and templates files into the binary by simply embeding folder using app.Embed

```go
app.Embed(staticDir *embed.FS, templateDir *embed.FS)
app.ServeLocalDir(dirPath, webPath string)
app.ServeEmbededDir(pathLocalDir string, embeded embed.FS, webPath string)
app.AddLocalTemplates(pathToDir string) error
app.AddEmbededTemplates(template_embed embed.FS,rootDir string) error
```

## Middlewares

#### Global server middlewares
```go
func main() {
	app := New()
	app.UseMiddlewares(
		middlewares.GZIP,
		middlewares.CORS,
		middlewares.CSRF,
		middlewares.LIMITER,
		middlewares.RECOVERY,
	)

	// GZIP nothing to do , just add it

	// LOGS /!\no need to add it using app.UseMiddlewares, instead you have a flag --logs that enable /logs
	// when logs middleware used, you will have a colored log for requests and also all logs from logger library displayed in the terminal and at /logs enabled for admin only
	// add the middleware like above and enjoy SSE logs in your browser not persisting if you ask

	// LIMITER 
	// if enabled , requester are blocked 5 minutes if make more then 50 request/s , you can change these values:
	middlewares.LIMITER_TOKENS=50
	middlewares.LIMITER_TIMEOUT=5*time.Minute

	// RECOVERY
	// will recover any error and log it, you can see it in console and also at /logs if LOGS middleware enabled

	// CORS
	// this is how to use CORS, it's applied globaly , but defined by the handler, all methods except GET of course
	app.POST(pattern string, handler kamux.Handler, allowed_origines ...string)
	app.POST("/users/post",func(c *kamux.Context) {
		// handle request here for domain.com and domain2.com and same origin
	},"domain.com","domain2.com")

	// CSRF
	// CSRF middleware get csrf_token from header, middleware will put it in cookies
	// this is a helper javascript function you can use to get cookie csrf and the set it globaly for all your request
	function getCookie(name) {
		var cookieValue = null;
		if (document.cookie && document.cookie != '') {
			var cookies = document.cookie.split(';');
			for (var i = 0; i < cookies.length; i++) {
				var cookie = cookies[i].trim();
				// Does this cookie string begin with the name we want?
				if (cookie.substring(0, name.length + 1) == (name + '=')) {
					cookieValue = decodeURIComponent(cookie.substring(name.length + 1));
					break;
				}
			}
		}
		return cookieValue;
	}
	let csrftoken = getCookie("csrf_token");

	// or a you have also a template function called csrf_token
	// you can use it in templates to render a hidden input of csrf_token
	<form id="form1">
		{{ csrf_token .request }}
	</form>
	// you can get it from the input from js and send it in headers
	// middleware csrf will try to find this header name:
	COOKIE NAME: 'csrf_token'
	HEADER NAME: 'X-CSRF-Token'



	app.Run()
}
```

## Handler middlewares

```go
// USAGE:
r.GET("/admin", middlewares.Admin(IndexView)) // will check from session cookie if user.is_admin is true
r.GET("/admin/login",middlewares.Auth(LoginView)) // will get session from cookies decrypt it and validate it
r.GET("/test",middlewares.BasicAuth(LoginView,"username","password"))
```

# ORM
###### i waited go1.18 and generics to make this package orm to keep performance at it's best with convenient way to query your data, even from multiple databases
## queries are cached using powerfull eventbus style that empty cache when changes in database may corrupt your data, so use it until you have a problem with it
###### to disable it : 
``` 
orm.UseCache=false
```

#### Let's start by ways to add new database:
```go
orm.NewDatabaseFromDSN(dbType,dbName string,dbDSN ...string) (error)
orm.NewDatabaseFromConnection(dbType,dbName string,conn *sql.DB) (error)
orm.GetConnection(dbName ...string) *sql.DB // get connection from dbName
orm.UseForAdmin(dbName string) // use specific database in admin panel if many
orm.GetDatabases() []DatabaseEntity // get list of databases available to your app
```

#### Utility database queries:
```go
orm.GetAllTables(dbName ...string) []string // if dbName not given, .env used instead to default the db to get all tables in the given db
orm.GetAllColumns(table string, dbName ...string) map[string]string // clear i think
orm.CreateUser(email,password string,isAdmin int, dbName ...string) error // password will be hashed using argon2
```

#### Migrations
##### you can migrate from a struct
```go
// execute AutoMigrate to migrate only, not if the table already exist in db, in this case you need only to link your model to the table using : orm.LinkModel
orm.AutoMigrate[T comparable](dbName, tableName string, debug ...bool) error // debug will print the query statement

// if table already exist you can use LinkModel[T comparable](to_table_name string, dbNames ...string)
orm.LinkModel[models.User]("users") // if dbNames empty , use default .env db

//Example:
type Bookmark struct {
	Id      uint   `orm:"autoinc"`
	UserId  int    `orm:"fk:users.id:cascade"`
	IsUsed	bool
	ToCheck string `orm:"size:50; notnull ;unique ; check:len(to_check) > 10"`
	Content string `orm:"text"`
	CreatedAt time.Time `orm:"now"`
}

// To migrate execute:
err := orm.AutoMigrate[Bookmark]("db", "bookmarks",true)

// will produce:
CREATE TABLE IF NOT EXISTS bookmarks (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
  user_id INTEGER, 
  is_used INTEGER NOT NULL CHECK (
    is_used IN (0, 1)
  ), 
  to_check VARCHAR(50) UNIQUE NOT NULL CHECK (
    length(to_check) > 10
  ), 
  content TEXT, 
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, 
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```
##### OR
##### using the shell, you can migrate a .sql file 'go run main.go shell'


## Queries and Sql Builder
#### to query, insert, update and delete using structs:
```go
orm.Model[T comparable]() *Builder[T]
orm.Database(dbName string) *Builder[T]
orm.Insert(model *T) (int, error)
orm.Set(query string, args ...any) (int, error)
orm.Delete() (int, error) // finisher
orm.Drop() (int, error) // finisher
orm.Select(columns ...string) *Builder[T]
orm.Where(query string, args ...any) *Builder[T]
orm.Query(query string, args ...any) *Builder[T]
orm.Limit(limit int) *Builder[T]
orm.Context(ctx context.Context) *Builder[T]
orm.Page(pageNumber int) *Builder[T]
orm.OrderBy(fields ...string) *Builder[T]
orm.Debug() *Builder[T] // print the query statement
orm.All() ([]T, error) // finisher
orm.One() (T, error) // finisher
```
#### Examples
```go
// if a model struct exist already in the database but not linked, you can't migrate it again, so you can link it using:
orm.LinkModel[models.User]("users")

// then you can query your data as models.User data
// you have 2 finisher : All and One

// Select and Pagination
orm.Model[models.User]().Select("email","uuid").OrderBy("-id").Limit(PAGINATION_PER).Page(1).All()

// INSERT
uuid,_ := orm.GenerateUUID()
hashedPass,_ := hash.GenerateHash("password")
orm.Model[models.User]().Insert(&models.User{
	Uuid: uuid,
	Email: "test@example.com",
	Password: hashedPass,
	IsAdmin: false,
	Image: "",
	CreatedAt: time.Now(),
})

//if using more than one db
orm.Database[models.User]("dbNameHere").Where("id = ? AND email = ?",1,"test@example.com").All() 

// where
orm.Model[models.User]().Where("id = ? AND email = ?",1,"test@example.com").One() 

// delete
orm.Model[models.User]().Where("id = ? AND email = ?",1,"test@example.com").Delete()

// drop table
orm.Model[models.User]().Drop()

// update
orm.Model[models.User]().Where("id = ?",1).Set("email = ?","new@example.com")

```
#### to query, insert, update and delete using map[string]any:
```go
orm.Table(tableName string) *BuilderM
orm.Database(dbName string) *BuilderM
orm.Select(columns ...string) *BuilderM
orm.Where(query string, args ...any) *BuilderM
orm.Query(query string, args ...any) *BuilderM
orm.Limit(limit int) *BuilderM
orm.Page(pageNumber int) *BuilderM
orm.OrderBy(fields ...string) *BuilderM
orm.Context(ctx context.Context) *BuilderM
orm.Debug() *BuilderM
orm.All() ([]map[string]any, error)
orm.One() (map[string]any, error)
orm.Insert(fields_comma_separated string, fields_values []any) (int, error)
orm.Set(query string, args ...any) (int, error)
orm.Delete() (int, error)
orm.Drop() (int, error)
```
#### Examples
```go
// for using maps , no need to link any model of course
// then you can query your data as models.User data
// you have 2 finisher : All and One for queries

// Select and Pagination
orm.Table("users").Select("email","uuid").OrderBy("-id").Limit(PAGINATION_PER).Page(1).All()

// INSERT
uuid,_ := orm.GenerateUUID()
hashedPass,_ := hash.GenerateHash("password")

orm.Table("users").Insert(
	"uuid,email,password,is_admin,created_at",
	uuid,
	"email@example.com",
	hashedPass,
	false,
	time.Now()
)

//if using more than one db
orm.Database("dbNameHere").Where("id = ? AND email = ?",1,"test@example.com").All() 

// where
orm.Table("users").Where("id = ? AND email = ?",1,"test@example.com").One() 

// delete
orm.Table("users").Where("id = ? AND email = ?",1,"test@example.com").Delete()

// drop table
orm.Table("users").Drop()

// update
orm.Table("users").Where("id = ?",1).Set("email = ?","new@example.com")

```



# SHELL
##### Very useful shell to explore, no need to install extra dependecies or binary, you can run:
```shell
go run main.go shell
go run main.go help
```

```shell
AVAILABLE COMMANDS:
[databases, use, tables, columns, migrate, createsuperuser, 
createuser, getall, get, drop, delete, clear/cls, q/quit/exit, help/commands]
  'databases':
	  list all connected databases

  'use':
	  use a specific database

  'tables':
	  list all tables in database

  'columns':
	  list all columns of a table

  'migrate':
	  migrate initial users to env database

  'createsuperuser':
	  create a admin user

  'createuser':
	  create a regular user

  'getall':
	  get all rows given a table name

  'get':
	  get single row wher field equal_to

  'delete':
	  delete rows where field equal_to

  'drop':
	  drop a table given table name

  'clear/cls':
	  clear console
```

## Database Benchmarks
```shell
////////////////////////////////////// postgres without cache
BenchmarkGetAllS-4                  1472            723428 ns/op            5271 B/op         80 allocs/op
BenchmarkGetAllM-4                  1502            716418 ns/op            4912 B/op         85 allocs/op
BenchmarkGetRowS-4                   826           1474674 ns/op            2288 B/op         44 allocs/op
BenchmarkGetRowM-4                   848           1392919 ns/op            2216 B/op         44 allocs/op
BenchmarkGetAllTables-4             1176            940142 ns/op             592 B/op         20 allocs/op
BenchmarkGetAllColumns-4             417           2862546 ns/op            1456 B/op         46 allocs/op
////////////////////////////////////// postgres with cache
BenchmarkGetAllS-4               2825896               427.9 ns/op           208 B/op          2 allocs/op
BenchmarkGetAllM-4               6209617               188.9 ns/op            16 B/op          1 allocs/op
BenchmarkGetRowS-4               2191544               528.1 ns/op           240 B/op          4 allocs/op
BenchmarkGetRowM-4               3799377               305.5 ns/op            48 B/op          3 allocs/op
BenchmarkGetAllTables-4         76298504                21.41 ns/op            0 B/op          0 allocs/op
BenchmarkGetAllColumns-4        59004012                19.92 ns/op            0 B/op          0 allocs/op
///////////////////////////////////// mysql without cache
BenchmarkGetAllS-4                  1221            865469 ns/op            7152 B/op        162 allocs/op
BenchmarkGetAllM-4                  1484            843395 ns/op            8272 B/op        215 allocs/op
BenchmarkGetRowS-4                   427           3539007 ns/op            2368 B/op         48 allocs/op
BenchmarkGetRowM-4                   267           4481279 ns/op            2512 B/op         54 allocs/op
BenchmarkGetAllTables-4              771           1700035 ns/op             832 B/op         26 allocs/op
BenchmarkGetAllColumns-4             760           1537301 ns/op            1392 B/op         44 allocs/op
///////////////////////////////////// mysql with cache
BenchmarkGetAllS-4               2933072               414.5 ns/op           208 B/op          2 allocs/op
BenchmarkGetAllM-4               6704588               180.4 ns/op            16 B/op          1 allocs/op
BenchmarkGetRowS-4               2136634               545.4 ns/op           240 B/op          4 allocs/op
BenchmarkGetRowM-4               4111814               292.6 ns/op            48 B/op          3 allocs/op
BenchmarkGetAllTables-4         58835394                21.52 ns/op            0 B/op          0 allocs/op
BenchmarkGetAllColumns-4        59059225                19.99 ns/op            0 B/op          0 allocs/op
///////////////////////////////////// sqlite without cache
BenchmarkGetAllS-4                 13664             85506 ns/op            2056 B/op         62 allocs/op
BenchmarkGetAllS_GORM-4            10000            101665 ns/op            9547 B/op        155 allocs/op
BenchmarkGetAllM-4                 13747             83989 ns/op            1912 B/op         61 allocs/op
BenchmarkGetAllM_GORM-4            10000            107810 ns/op            8387 B/op        237 allocs/op
BenchmarkGetRowS-4                 12702             91958 ns/op            2192 B/op         67 allocs/op
BenchmarkGetRowM-4                 13256             89095 ns/op            2048 B/op         66 allocs/op
BenchmarkGetAllTables-4            14264             83939 ns/op             672 B/op         32 allocs/op
BenchmarkGetAllColumns-4           15236             79498 ns/op            1760 B/op         99 allocs/op
///////////////////////////////////// sqlite with cache
BenchmarkGetAllS-4               2951642               399.5 ns/op           208 B/op          2 allocs/op
BenchmarkGetAllM-4               6537204               177.2 ns/op            16 B/op          1 allocs/op
BenchmarkGetRowS-4               2248524               531.4 ns/op           240 B/op          4 allocs/op
BenchmarkGetRowM-4               4084453               287.9 ns/op            48 B/op          3 allocs/op
BenchmarkGetAllTables-4         52592826                20.39 ns/op            0 B/op          0 allocs/op
BenchmarkGetAllColumns-4        64293176                20.87 ns/op            0 B/op          0 allocs/op

```


## OpenApi documentation ready if enabled using flags or settings vars
```bash
// running the app using flag --docs
go run main.go --docs
// go to /docs
```
## to edit the documentation, you have a docs package 
```go

doc := docs.New()

doc.AddModel("User",docs.Model{
	Type: "object",
	RequiredFields: []string{"email","password"},
	Properties: map[string]docs.Property{
		"email":{
			Required: true,
			Type: "string",
			Example: "example@xyz.com",
		},
		"password":{
			Required: true,
			Type: "string",
			Example: "************",
			Format: "password",
		},
	},
})
doc.AddPath("/admin/login","post",docs.Path{
	Tags: []string{"Auth"},
	Summary: "login post request",
	OperationId: "login-post",
	Description: "Login Post Request",
	Requestbody: docs.RequestBody{
		Description: "email and password for login",
		Required: true,
		Content: map[string]docs.ContentType{
			"application/json":{
				Schema: docs.Schema{
					Ref: "#/components/schemas/User",
				},
			},
		},
	},
	Responses: map[string]docs.Response{
		"404":{Description: "NOT FOUND"},
		"403":{Description: "WRONG PASSWORD"},
		"500":{Description: "INTERNAL SERVER ERROR"},
		"200":{Description: "OK"},
	},
	Consumes: []string{"application/json"},
	Produces: []string{"application/json"},
})
doc.Save()
```

## Encryption
```go
// AES Encrypt use SECRET in .env
encryptor.Encrypt(data string) (string,error)
encryptor.Decrypt(data string) (string,error)
```

## Hashing
```go
// Argon2 hashing
hash.GenerateHash(password string) (string, error)
hash.ComparePasswordToHash(password, hash string) (bool, error)
```

## Env Loader
```go
envloader.Load(files ...string) error
envloader.LoadToMap(files ...string) (map[string]string,error)
```

## Eventbus Internal
```go
eventbus.Subscribe("any_topic",func(data map[string]string) {
	...
})

eventbus.Publish("any_topic", map[string]string{
	"type":     "update",
	"table":    b.tableName,
	"database": b.database,
})
```

## LOGS
```bash
go run main.go --logs

will enable:
	- /logs
```

## PPROF official golang profiling tools
```bash
go run main.go --profiler

will enable:
	- /debug/pprof/profile
	- /debug/pprof/heap
	- /debug/pprof/trace
```

## Prometheus monitoring
```bash
go run main.go --monitoring

will enable:
	- /metrics
```

## Build single binary with all static and html files
```bash
#you nedd to set at .env to true: 
EMBED_STATIC=true
EMBED_TEMPLATES=true
```
#### Then

```go
//go:embed assets/static
var Static embed.FS
//go:embed assets/templates
var Templates embed.FS

func main() {
	app := New()
	app.UseMiddlewares(middlewares.GZIP)
	app.Embed(&Static,&Templates)
	app.Run()
}

```

#### Then
```bash
go build
```


## 🔗 Links
[![portfolio](https://img.shields.io/badge/my_portfolio-000?style=for-the-badge&logo=ko-fi&logoColor=white)](https://kamalshkeir.github.io/)
[![linkedin](https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/kamal-shkeir/)


# Licence
Licence [BSD-3](./LICENSE)