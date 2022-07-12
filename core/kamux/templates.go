package kamux

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kamalshkeir/kago/core/orm"
	"github.com/kamalshkeir/kago/core/settings"
	"github.com/kamalshkeir/kago/core/utils"
	"github.com/kamalshkeir/kago/core/utils/logger"
)

var allTemplates = template.New("")

// InitTemplatesAndAssets init templates from a folder and download admin skeleton html files
func (router *Router) InitTemplatesAndAssets() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		router.cloneTemplatesAndStatic()
		wg.Done()
	}()
	go func() {
		router.initAdminDocsAndStaticURLS()
		wg.Done()
	}()
	wg.Wait()
	if settings.GlobalConfig.EmbedTemplates {
		findAndParseEmbededTemplates(Templates,"assets/templates",functions)
	} else {
		//local
		findAndParseTemplates("assets/templates",functions)
	}
}

func (router *Router) NewFuncMap(funcName string, function any) {
	if _,ok := functions[funcName];ok {
		logger.Error("unable to add",funcName,",already exist")
	} else {
		functions[funcName]=function
	}
}

func (router *Router) cloneTemplatesAndStatic()  {
    var generated bool
    repo_name := "kamal_assets"
    new_name := "assets"
    if _,err := os.Stat(new_name);err != nil && !settings.GlobalConfig.EmbedStatic && !settings.GlobalConfig.EmbedTemplates {
        // if not generated
        cmd := exec.Command("git", "clone", "https://github.com/kamalshkeir/"+repo_name)
        err := cmd.Run()
		if logger.CheckError(err) {return}
		err = os.RemoveAll(repo_name+"/.git")
		if err != nil {
			logger.Error("unable to delete",repo_name+"/.git :",err)
		}
        err = os.Rename(repo_name,new_name)
		if err != nil {
			logger.Error("unable to rename",repo_name,err)
		}
		generated=true
    } 

	tables := orm.GetAllTables()
	found := false
	for _,t := range tables {
		if t == "users" {
			found=true
		}
	}
	if !found {
		fmt.Printf(logger.Blue,"initial models migrated")
		err := orm.Migrate()
		if !logger.CheckError(err) {
			fmt.Printf(logger.Blue,"you can run 'go run main.go createsuperuser'")
		}
		os.Exit(0)
	} 
	
    if generated && !found {
		// migrate initial models
		err := orm.Migrate()
		logger.CheckError(err)
		fmt.Printf(logger.Green,"assets generated, initital models migrated")
		fmt.Printf(logger.Blue,"exec: go run main.go createsuperuser to create your admin account")
        os.Exit(0)
    } else if generated {
		fmt.Printf(logger.Green,"assets generated")
	}
}


func (router *Router) initAdminDocsAndStaticURLS() {
    // PROFILER
	if settings.GlobalConfig.Profiler {
		router.Get("^/debug/pprof/?heap", func(c *Context) { pprof.Index(c.ResponseWriter, c.Request) })
		router.Get("^/debug/pprof/profile", func(c *Context) { pprof.Profile(c.ResponseWriter, c.Request) })
		router.Get("^/debug/pprof/trace", func(c *Context) { pprof.Trace(c.ResponseWriter, c.Request) })
	}
	// STATIC
	router.Get("/favicon.ico", func(c *Context) { http.NotFoundHandler().ServeHTTP(c.ResponseWriter, c.Request) })
	if settings.GlobalConfig.EmbedStatic {
		//EMBED STATIC
		staticFs := fs.FS(Static)
		content,err := fs.Sub(staticFs,"assets")
		logger.CheckError(err)
		static_root := http.FileServer(http.FS(content))
		router.Get(`^/static/`, func(c *Context) { static_root.ServeHTTP(c.ResponseWriter, c.Request) })
		if settings.GlobalConfig.Docs {
			swagger_dir, _ := fs.Sub(content, "static/docs")
			docs_root := http.FileServer(http.FS(swagger_dir))
			router.Get(`^/docs/`, func(c *Context) {
				http.StripPrefix("/docs/", docs_root).ServeHTTP(c.ResponseWriter, c.Request)
			})
			router.Get("/docs",func(c *Context) {c.Redirect("/docs/",http.StatusPermanentRedirect)})
		}
	} else {
		// LOCAL STATIC
		static_root := http.FileServer(http.Dir("./assets/static"))
		router.Get(`^/static/`, func(c *Context) {
			http.StripPrefix("/static/", static_root).ServeHTTP(c.ResponseWriter, c.Request)
		})
		if settings.GlobalConfig.Docs {
			docs_root := http.FileServer(http.Dir("./assets/static/docs"))
			
			router.Get(`^/docs/`, func(c *Context) {
				http.StripPrefix("/docs/", docs_root).ServeHTTP(c.ResponseWriter, c.Request)
			})
			router.Get("/docs",func(c *Context) {c.Redirect("/docs/",http.StatusPermanentRedirect)})
		}
	}
	// MEDIA
	media_root := http.FileServer(http.Dir("./media"))
	router.Get(`^/media/`, func(c *Context) {
		http.StripPrefix("/media/", media_root).ServeHTTP(c.ResponseWriter, c.Request)
	})
}

/* FUNC MAPS */
var functions = template.FuncMap{
	"isBool": func(something any) bool {
		res := false
		switch v := something.(type) {
		case string:
			if v == "true" || v == "1" || v == "false" || v == "0" {
				res=true
			}
		case int:
			if v == 1 || v == 0 {
				res = true
			}
		case int64:
			if int(v) == 1 || v == 0 {
				res = true
			}
		case bool:
			res=true
		case uint64:
			if int(v) == 1 || v == 0 {
				res = true
			}
		default:
			res=false
		}
		return res
	},
	"isTrue": func(something any) bool {
		res := false
		switch v := something.(type) {
		case string:
			if v == "true" || v == "1" {
				res=true
			}
		case int:
			if v == 1 {
				res = true
			}
		case int64:
			if int(v) == 1 {
				res = true
			}
		case uint64:
			if int(v) == 1 {
				res = true
			}
		case bool:
			res=v
		default:
			res=false
		}
		return res
	},
	"contains": func(str string, substrings ...string) bool {
		for _, substr := range substrings {
			if strings.Contains(strings.ToLower(str), substr) {
				return true
			}
		}
		return false
	},
	"startWith": func(str string, substrings ...string) bool {
		for _, substr := range substrings {
			if strings.HasPrefix(strings.ToLower(str), substr) {
				return true
			}
		}
		return false
	},
	"finishWith": func(str string, substrings ...string) bool {
		for _, substr := range substrings {
			if strings.HasSuffix(strings.ToLower(str), substr) {
				return true
			}
		}
		return false
	},
	"generateUUID": func() template.HTML {
		uuid, _ := utils.GenerateUUID()
		return template.HTML(uuid)
	},
	"inc": func(i int) int {
		return i + 1
	},
	"safe": func(str string) template.HTML {
		return template.HTML(str)
	},
	"timeFormat":func (t any) string {
		valueToReturn := ""
		switch v := t.(type) {
		case time.Time:
			if !v.IsZero() {
				valueToReturn = v.Format("2006-01-02T15:04")
			} else {
				valueToReturn = time.Now().Format("2006-01-02T15:04")
			}
		case string:
			if len(v) >= len("2006-01-02T15:04") && strings.Contains(v[:13],"T") {
				p,err := time.Parse("2006-01-02T15:04",v)
				if logger.CheckError(err) {
					valueToReturn = time.Now().Format("2006-01-02T15:04")
				} else {
					valueToReturn = p.Format("2006-01-02T15:04")
				}
			} else {
				if len(v) >= 16 {
					p,err := time.Parse("2006-01-02 15:04",v[:16])
					if logger.CheckError(err) {
						valueToReturn = time.Now().Format("2006-01-02T15:04")
					} else {
						valueToReturn = p.Format("2006-01-02T15:04")
					}	
				} 
			}
		default:
			logger.Error("type of",t,"is not handled,type is:",v)
			valueToReturn = ""
		}
		return valueToReturn
	},
	"truncate": func(str any,size int) any {
		switch v := str.(type) {
		case string:
			if len(v) > size {
				return v[:size]
			} else {
				return v
			}
		default:
			return v
		}
	},
	"csrf_token":func (r *http.Request) template.HTML {
		csrf,_ := r.Cookie("csrf_token")
		if csrf != nil {
			return template.HTML(fmt.Sprintf("   <input type=\"hidden\" id=\"csrf_token\" value=\"%s\">   ",csrf.Value))
		} else {
			return template.HTML("")
		}
	},
}


func findAndParseTemplates(rootDir string, funcMap template.FuncMap) error {
    cleanRoot := filepath.ToSlash(rootDir)
    pfx := len(cleanRoot)+1

    err := filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
        if !info.IsDir() && strings.HasSuffix(path, ".html") {
            if e1 != nil {
                return e1
            }

            b, e2 := ioutil.ReadFile(path)
            if e2 != nil {
                return e2
            }

            name := filepath.ToSlash(path[pfx:])
            t := allTemplates.New(name).Funcs(funcMap)
            _, e2 = t.Parse(string(b))
            if e2 != nil {
                return e2
            }
        }

        return nil
    })

    return err
}

func findAndParseEmbededTemplates(template_embed embed.FS,rootDir string, funcMap template.FuncMap) error {
	cleanRoot := filepath.ToSlash(rootDir)
    pfx := len(cleanRoot)+1
    err := fs.WalkDir(template_embed,cleanRoot,func(path string, info fs.DirEntry, e1 error) error {
		logger.CheckError(e1)
        if !info.IsDir() && strings.HasSuffix(path, ".html") {
            if logger.CheckError(e1) {
                return e1
            }
			b,e2 := template_embed.ReadFile(path)
			if logger.CheckError(e2) {
                return e2
            }

            name := filepath.ToSlash(path[pfx:])
            t := allTemplates.New(name).Funcs(funcMap)
            _, e3 := t.Parse(string(b))
            if logger.CheckError(e3) {
                return e2
            }
        }

        return nil
    })
	

    return err
}
