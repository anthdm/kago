package orm

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/kamalshkeir/kago/core/admin/models"
	"github.com/kamalshkeir/kago/core/settings"
	"github.com/kamalshkeir/kago/core/utils"
	"github.com/kamalshkeir/kago/core/utils/logger"
)

func Migrate() error {
	err := AutoMigrate[models.User](settings.GlobalConfig.DbName,"users",settings.GlobalConfig.DbType)
	if logger.CheckError(err) {
		return err
	}
	switch settings.GlobalConfig.DbType {
	case "postgres":
		_, err := GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version INTEGER,
			name VARCHAR(100) UNIQUE NOT NULL,
			up_path VARCHAR(100) DEFAULT '',
			down_path VARCHAR(100) DEFAULT '',
			executed INTEGER NOT NULL CHECK (executed IN (0, 1)) DEFAULT 0,
			created_at TIMESTAMP with time zone NOT NULL DEFAULT (now())
		)`)
		if logger.CheckError(err) {
			return err
		}
		/* USERS */
		/* _, err = GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				uuid VARCHAR(50) DEFAULT '',
				email VARCHAR(50) UNIQUE NOT NULL,
				password VARCHAR(255) NOT NULL,
				is_admin INTEGER NOT NULL CHECK (is_admin IN (0, 1)) DEFAULT 0,
				image VARCHAR(255) DEFAULT '',
				created_at TIMESTAMP with time zone NOT NULL DEFAULT (now())
		)`)
		if logger.CheckError(err) {
			return err
		} */
		return nil
	case "mysql":
		_, err := GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id MEDIUMINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			version INTEGER,
			name VARCHAR(100) UNIQUE NOT NULL,
			up_path VARCHAR(100) DEFAULT '',
			down_path VARCHAR(100) DEFAULT '',
			executed INTEGER NOT NULL CHECK (executed IN (0, 1)) DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
		if logger.CheckError(err) {
			return err
		}
		/* USERS */
		/* _, err = GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS users (
				id MEDIUMINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
				uuid VARCHAR(50) DEFAULT '',
				email VARCHAR(50) UNIQUE NOT NULL,
				password VARCHAR(255) NOT NULL,
				is_admin INTEGER NOT NULL CHECK (is_admin IN (0, 1)) DEFAULT 0,
				image VARCHAR(255) DEFAULT '',
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
		if logger.CheckError(err) {
			return err
		} */
		return nil
	case "sqlite":
		_, err := GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER,
			name VARCHAR(100) UNIQUE NOT NULL,
			up_path VARCHAR(100) DEFAULT '',
			down_path VARCHAR(100) DEFAULT '',
			executed INTEGER NOT NULL CHECK (executed IN (0, 1)),
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
		if logger.CheckError(err) {
			return err
		}
		/* USERS */
		/* _, err = GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				uuid VARCHAR(50) DEFAULT '',
				email VARCHAR(50) UNIQUE NOT NULL,
				password VARCHAR(255) NOT NULL,
				is_admin INTEGER NOT NULL CHECK (is_admin IN (0, 1)),
				image VARCHAR(255) DEFAULT '',
				created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
		if logger.CheckError(err) {
			return err
		} */

		return nil
	}
	return nil
}


func AutoMigrate[T comparable](dbName,tableName,dialect string, debug ...bool) error {
	s := reflect.ValueOf(new(T)).Elem()
	typeOfT := s.Type()
	mFieldName_Type := map[string]string{}
	mFieldName_Tags := map[string][]string{}
	cols := []string{}
	
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fname := typeOfT.Field(i).Name
		fname = ToSnakeCase(fname)
		ftype := f.Type()
		cols = append(cols, fname)
		mFieldName_Type[fname]=ftype.Name()
		if ftag,ok := typeOfT.Field(i).Tag.Lookup("orm");ok {
			tags := strings.Split(ftag,";")
			mFieldName_Tags[fname]=tags
		} 
	}

	res := map[string]string{}
	fkeys := []string{}
	utils.ReverseSlice(cols)
	for _,fName := range cols {
		if ty,ok := mFieldName_Type[fName];ok {
			switch ty  {
			case "int","uint","int64","uint64","int32","uint32":
				handleMigrationInt(dialect, fName,ty,&mFieldName_Tags,&fkeys,&res)
			case "bool":
				handleMigrationBool(dialect, fName,ty,&mFieldName_Tags,&fkeys,&res)				
			case "string":
				handleMigrationString(dialect, fName,ty,&mFieldName_Tags,&fkeys,&res)
			case "float64","float32":
				handleMigrationFloat(dialect, fName,ty,&mFieldName_Tags,&fkeys,&res)
			case "Time":
				handleMigrationTime(dialect, fName,ty,&mFieldName_Tags,&fkeys,&res)
			default:
				logger.Error(fName,"of type",ty,"not handled")
			}
		}	
	}
	
	statement := prepareCreateStatement(tableName,res,fkeys,cols)
	if len(debug) > 0 && debug[0] {
		logger.Debug("statement:",statement)
	}
	if conn,ok := mDbNameConnection[dbName];ok {
		c,cancel := context.WithTimeout(context.TODO(),3*time.Second)
		defer cancel()
		res,err := conn.ExecContext(c,statement)
		if err != nil {
			logger.Info(statement)
			return err
		}
		_, err = res.RowsAffected()
		if err != nil {
			return err
		}
		tables := GetAllTables(dbName)
		if len(tables) > 0 {
			for _,t := range tables {
				if t == tableName {
					LinkModel[T](tableName)
				}
			}
		}
	} else {
		logger.Info(mDbNameConnection)
		return errors.New("no connection found for "+dbName)
	}
	
	return nil
}

func handleMigrationInt(dialect, fName,ty string,mFieldName_Tags *map[string][]string,fkeys *[]string,res *map[string]string) {
	primary,autoinc,unique,notnull,defaultt:="","","","",""			
	tags := (*mFieldName_Tags)[fName]
	for _,tag := range tags {
		switch tag {
		case "pk":
			primary=" PRIMARY KEY"
		case "autoinc":
			switch dialect {
			case "sqlite","":
				autoinc="INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT"
			case "postgres":
				autoinc="SERIAL NOT NULL PRIMARY KEY"
			case "mysql":
				autoinc="MEDIUMINT NOT NULL PRIMARY KEY AUTO_INCREMENT"
			default:
				logger.Error("dialect can be sqlite, postgres or mysql only, not ",dialect)
			}
		case "unique":
			unique=" UNIQUE"
		case "notnull":
			notnull="NOT NULL"
		default:
			if strings.Contains(tag,":") {
				sp := strings.Split(tag,":")
				switch sp[0] {
				case "default":
					defaultt=sp[1]
				case "fk":
					ref := strings.Split(sp[1],".")
					if len(ref) == 2 {
						fkey := "FOREIGN KEY ("+fName+") REFERENCES "+ref[0]+"("+ref[1]+")"
						if len(sp) > 2 {
							if sp[2] == "cascade" {
								fkey += " ON DELETE CASCADE"
							} else if sp[2] == "donothing" || sp[2] == "noaction" {
								fkey += " ON DELETE NO ACTION"
							}
						}
						*fkeys = append(*fkeys, fkey)
					} else {
						logger.Error("wtf ?, it should be fk:users.id:cascade/donothing")
					}
					
				default:
					logger.Error("not handled",sp[0],"for",tag,",field:",fName)
				}
				
			} else {
				logger.Error("tag",tag,"not handled for",fName,"of type",ty)
			}	
		}
	}

	if autoinc != "" {
		// integer auto increment
		(*res)[fName] = autoinc
	} else {
		// integer normal
		(*res)[fName]="INTEGER"
		if primary != "" {
			(*res)[fName]+=primary
		}
		if unique != "" {
			(*res)[fName]+=unique
		}
		if notnull != "" {
			(*res)[fName]+=notnull
		}
		if defaultt != "" {
			(*res)[fName]+=" DEFAULT " + defaultt
		}
	}
}

func handleMigrationBool(dialect, fName,ty string,mFieldName_Tags *map[string][]string,fkeys *[]string,res *map[string]string) {
	defaultt := ""
	(*res)[fName]="INTEGER NOT NULL CHECK ("+fName+" IN (0, 1))"
	tags := (*mFieldName_Tags)[fName]
	for _,tag := range tags {
		if strings.Contains(tag,":") {
			sp := strings.Split(tag,":")
			switch sp[0] {
			case "default":
				if sp[1] == "true" || sp[1] == "1" {
					defaultt = " DEFAULT 1" 
				} else {
					defaultt = " DEFAULT 0"
				}
			case "fk":
				ref := strings.Split(sp[1],".")
				if len(ref) == 2 {
					fkey := "FOREIGN KEY(\""+fName+"\") REFERENCES "+ref[0]+"(\""+ref[1]+"\")"
					if len(sp) > 2 {
						if sp[2] == "cascade" {
							fkey += " ON DELETE CASCADE"
						} else if sp[2] == "donothing" || sp[2] == "noaction" {
							fkey += " ON DELETE NO ACTION"
						}
					}
					*fkeys = append(*fkeys, fkey)
				} else {
					logger.Error("wtf ?, it should be fk:users.id:cascade/donothing")
				}						
			default:
				logger.Error("not handled",sp[0],"for",tag,",field:",fName)
			}					
		} else {
			logger.Error("tag",tag,"not handled for",fName,"of type",ty)
		}
		if defaultt != "" {
			(*res)[fName] += defaultt
		}
	}	
}

func handleMigrationString(dialect, fName,ty string,mFieldName_Tags *map[string][]string,fkeys *[]string,res *map[string]string) {
	unique,notnull,text,defaultt,size:="","","","",""				
				tags := (*mFieldName_Tags)[fName]
				for _,tag := range tags {
					switch tag {
					case "unique":
						unique=" UNIQUE"
					case "text":
						text=" TEXT"
					case "notnull":
						notnull=" NOT NULL"				
					default:
						if strings.Contains(tag,":") {
							sp := strings.Split(tag,":")
							switch sp[0] {
							case "default":
								if sp[1] == "" {
									defaultt=" DEFAULT "+sp[1]
								}								
							case "fk":
								ref := strings.Split(sp[1],".")
								if len(ref) == 2 {
									fkey := "FOREIGN KEY(\""+fName+"\") REFERENCES "+ref[0]+"(\""+ref[1]+"\")"
									if len(sp) > 2 {
										if sp[2] == "cascade" {
											fkey += " ON DELETE CASCADE"
										} else if sp[2] == "donothing" || sp[2] == "noaction" {
											fkey += " ON DELETE NO ACTION"
										}
									}
									*fkeys = append(*fkeys, fkey)
								} else {
									logger.Error("foreign key should be like fk:table.column:[cascade/donothing]")
								}	
							case "size":
								sp := strings.Split(tag,":")
								if sp[0] == "size" {
									size=sp[1]
								}				
							default:
								logger.Error("not handled",sp[0],"for",tag,",field:",fName)
							}					
						} else {
							logger.Error("tag",tag,"not handled for",fName,"of type",ty)
						}
					}
				}

				if text != "" {
					(*res)[fName]=text
				} else {
					if size != "" {
						(*res)[fName]="VARCHAR("+size+")"
					} else {
						(*res)[fName]="VARCHAR(255)"
					}
				}

				if unique != "" {
					(*res)[fName]+=unique
				}
				if notnull != "" {
					(*res)[fName]+=notnull
				}
				if defaultt !=  "" {
					(*res)[fName]+=defaultt
				}
}

func handleMigrationFloat(dialect, fName,ty string,mFieldName_Tags *map[string][]string,fkeys *[]string,res *map[string]string) {
	defaultt := ""
	(*res)[fName]="DECIMAL(5,2)"
	tags := (*mFieldName_Tags)[fName]
	for _,tag := range tags {
		if strings.Contains(tag,":") {
			sp := strings.Split(tag,":")
			switch sp[0] {
			case "default":
				if sp[1] != "" {
					defaultt = " DEFAULT "+ sp[1]
				}
			case "fk":
				ref := strings.Split(sp[1],".")
				if len(ref) == 2 {
					fkey := "FOREIGN KEY(\""+fName+"\") REFERENCES "+ref[0]+"(\""+ref[1]+"\")"
					if len(sp) > 2 {
						if sp[2] == "cascade" {
							fkey += " ON DELETE CASCADE"
						} else if sp[2] == "donothing" || sp[2] == "noaction" {
							fkey += " ON DELETE NO ACTION"
						}
					}
					*fkeys = append(*fkeys, fkey)
				} else {
					logger.Error("wtf ?, it should be fk:users.id:cascade/donothing")
				}						
			default:
				logger.Error("not handled",sp[0],"for",tag,",field:",fName)
			}					
		} else {
			logger.Error("tag",tag,"not handled for",fName,"of type",ty)
		}
		if defaultt != "" {
			(*res)[fName] += defaultt
		}
	}
}
		
func handleMigrationTime(dialect, fName,ty string,mFieldName_Tags *map[string][]string,fkeys *[]string,res *map[string]string) {
	defaultt,notnull := "",""
	tags := (*mFieldName_Tags)[fName]
	for _,tag := range tags {
		if tag == "now" {
			switch dialect {
			case "sqlite","":
				defaultt = "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP"
			case "postgres":
				defaultt = "TIMESTAMP with time zone NOT NULL DEFAULT (now())"
			case "mysql":
				defaultt = "TIMESTAMP with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP"
			default:
				logger.Error("not handled Time for ",fName,ty)
			}
		} else if tag == "notnull" {
			notnull = " NOT NULL"
		} else if strings.Contains(tag,":") {
			sp := strings.Split(tag,":") 
			if sp[0] == "fk" {
				ref := strings.Split(sp[1],".")
				if len(ref) == 2 {
					fkey := "FOREIGN KEY(\""+fName+"\") REFERENCES "+ref[0]+"(\""+ref[1]+"\")"
					if len(sp) > 2 {
						if sp[2] == "cascade" {
							fkey += " ON DELETE CASCADE"
						} else if sp[2] == "donothing" || sp[2] == "noaction" {
							fkey += " ON DELETE NO ACTION"
						}
					}
					*fkeys = append(*fkeys, fkey)
				} else {
					logger.Error("wtf ?, it should be fk:users.id:cascade/donothing")
				}
			} else if sp[0] == "default" {
				if sp[1] == "" {
					switch dialect {
					case "sqlite","":
						defaultt = "TEXT NOT NULL DEFAULT "+sp[1]
					case "postgres":
						defaultt = "TIMESTAMP with time zone NOT NULL DEFAULT "+sp[1]
					case "mysql":
						defaultt = "TIMESTAMP with time zone NOT NULL DEFAULT "+sp[1]
					default:
						logger.Error("default for field",fName,"not handled")
					}
				}	
			}
		} else {
			logger.Error("tag",tag,"not handled")
		}
	}
	if defaultt != "" {
		(*res)[fName]=defaultt
	} else {
		if dialect == "" || dialect == "sqlite" {
			(*res)[fName]="TEXT"
		} else {
			(*res)[fName]="TIMESTAMP with time zone"
		}
		if notnull != "" {
			(*res)[fName] += notnull
		}
	}
}

		
func prepareCreateStatement(tbName string,fields map[string]string,fkeys,cols []string) string {
	utils.ReverseSlice(cols)
	st := "CREATE TABLE IF NOT EXISTS "
	st += tbName + " ("
	for i,col := range cols {
		fName := col
		fType := fields[col]
		reste := ","
		if i == len(fields)-1 {
			reste = ""
		}
		st += fName + " " + fType + reste 
	}
	if len(fkeys) > 0 {
		st += ","
	}
	for i,k := range fkeys {
		st += k
		if i < len(fkeys)-1 {
			st += ","
		}
	}
	st += ");"
	return st
}