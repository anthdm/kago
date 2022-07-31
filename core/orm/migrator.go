package orm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kamalshkeir/kago/core/admin/models"
	"github.com/kamalshkeir/kago/core/settings"
	"github.com/kamalshkeir/kago/core/utils/logger"
)

func Migrate() error {
	err := AutoMigrate[models.User]("users",settings.GlobalConfig.DbName)
	if logger.CheckError(err) {
		return err
	}
	return nil
}

func autoMigrate[T comparable](db *DatabaseEntity, tableName string, debug ...bool) error {
	dialect := db.Dialect
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
		cols = append(cols,fname)
		mFieldName_Type[fname] = ftype.Name()
		if ftag, ok := typeOfT.Field(i).Tag.Lookup("orm"); ok {
			tags := strings.Split(ftag, ";")
			for i := range tags {
				tags[i]=strings.TrimSpace(tags[i])
			}
			mFieldName_Tags[fname] =  tags
		}
	}
	res := map[string]string{}
	fkeys := []string{}
	for _, fName := range cols {
		if ty, ok := mFieldName_Type[fName]; ok {
			switch ty {
			case "int", "uint", "int64", "uint64", "int32", "uint32":
				handleMigrationInt(dialect, fName, ty, &mFieldName_Tags, &fkeys, &res)
			case "bool":
				handleMigrationBool(dialect, fName, ty, &mFieldName_Tags, &fkeys, &res)
			case "string":
				handleMigrationString(dialect, fName, ty, &mFieldName_Tags, &fkeys, &res)
			case "float64", "float32":
				handleMigrationFloat(dialect, fName, ty, &mFieldName_Tags, &fkeys, &res)
			case "Time":
				handleMigrationTime(dialect, fName, ty, &mFieldName_Tags, &fkeys, &res)
			default:
				logger.Error(fName, "of type", ty, "not handled")
			}
		}
	}
	statement := prepareCreateStatement(tableName, res, fkeys, cols,db,mFieldName_Tags)
	tbFound := false
	for _,t := range db.Tables {
		if t.Name == tableName {
			tbFound=true
			if len(t.Columns) == 0 {t.Columns=cols}
			if len(t.Tags) == 0 {t.Tags=mFieldName_Tags}
			if len(t.Types) == 0 {t.Types=mFieldName_Type}
		}
	}
	if !tbFound {
		db.Tables = append(db.Tables, TableEntity{
			Name: tableName,
			Columns: cols,
			Tags:mFieldName_Tags,
			ModelTypes: mFieldName_Type,
		})
	}
	if len(debug) > 0 && debug[0] {
		fmt.Printf(logger.Blue,"statement: "+ statement)
	}
	
	c, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	ress, err := db.Conn.ExecContext(c, statement)
	if err != nil {
		return err
	}
	_, err = ress.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func AutoMigrate[T comparable](tableName string, dbName ...string) error {
	if _,ok := mModelTablename[*new(T)];!ok {
		mModelTablename[*new(T)]=tableName
	}
	var db *DatabaseEntity
	var err error
	dbname := ""
	if len(dbName) > 0 {
		dbname = dbName[0]
		db,err = GetDatabase(dbname)
		if err != nil || db == nil {
			return errors.New("database not found")
		}
	} else {
		dbname = settings.GlobalConfig.DbName
		db,err = GetDatabase(dbname)
		if err != nil || db == nil {
			return errors.New("database not found")
		}
	}
	
	tbFoundDB := false
	tables := GetAllTables(dbname)
	for _, t := range tables {
		if t == tableName {
			tbFoundDB=true
		}
	}
	
	tbFoundLocal := false
	if len(db.Tables) == 0 {
		if tbFoundDB {
			// found db not local
			linkModel[T](tableName,db)
			return nil
		} else {
			// not db and not local
			err := autoMigrate[T](db,tableName,false)
			if logger.CheckError(err) {
				return err
			}
			return nil
		}
	} else {
		// db have tables
		for _,t := range db.Tables {
			if t.Name == tableName {
				tbFoundLocal=true
			}
		}
	} 	
	if !tbFoundLocal {
		if tbFoundDB {
			linkModel[T](tableName,db)
			return nil
		} else {
			err := autoMigrate[T](db,tableName,false)
			if logger.CheckError(err) {
				return err
			}
		}
	} 

	return nil
}

func handleMigrationInt(dialect, fName, ty string, mFieldName_Tags *map[string][]string, fkeys *[]string, res *map[string]string) {
	primary,index, autoinc, notnull, defaultt, check := "", "", "", "", "", ""
	tags := (*mFieldName_Tags)[fName]
	for _, tag := range tags {
		switch tag {
		case "pk":
			primary = " PRIMARY KEY"
		case "autoinc":
			switch dialect {
			case "sqlite", "":
				autoinc = "INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT"
			case "postgres":
				autoinc = "SERIAL NOT NULL PRIMARY KEY"
			case "mysql":
				autoinc = "INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT"
			default:
				logger.Error("dialect can be sqlite, postgres or mysql only, not ", dialect)
			}
		case "notnull":
			notnull = "NOT NULL"
		default:
			if strings.Contains(tag, ":") {
				sp := strings.Split(tag, ":")
				switch sp[0] {
				case "default":
					defaultt = " DEFAULT " + sp[1]
				case "fk":
					ref := strings.Split(sp[1], ".")
					if len(ref) == 2 {
						fkey := "FOREIGN KEY (" + fName + ") REFERENCES " + ref[0] + "(" + ref[1] + ")"
						if len(sp) > 2 {
							if sp[2] == "cascade" {
								fkey += " ON DELETE CASCADE"
							} else if sp[2] == "donothing" || sp[2] == "noaction" {
								fkey += " ON DELETE NO ACTION"
							}
						}
						*fkeys = append(*fkeys, fkey)
					} else {
						logger.Error("allowed options cascade/donothing/noaction")
					}
				case "check":
					if strings.Contains(strings.ToLower(sp[1]), "len") {
						switch dialect {
						case "sqlite", "":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "length", 1)
						case "postgres", "mysql":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "char_length", 1)
						default:
							logger.Error("check not handled for dialect:", dialect)
						}
					}
					check = " CHECK (" + sp[1] + ")"
				default:
					logger.Error("not handled", sp[0], "for", tag, ",field:", fName)
				}

			} else {
				logger.Error("tag", tag, "not handled for", fName, "of type", ty)
			}
		}
	}

	if autoinc != "" {
		// integer auto increment
		(*res)[fName] = autoinc
	} else {
		// integer normal
		(*res)[fName] = "INTEGER"
		if primary != "" {
			(*res)[fName] += primary
		} else {
			if notnull != "" {
				(*res)[fName] += notnull
			}
		}
		if index != "" {
			(*res)[fName] += index
		}
		if defaultt != "" {
			(*res)[fName] += defaultt
		}
		if check != "" {
			(*res)[fName] += check
		}
	}
}

func handleMigrationBool(_, fName, ty string, mFieldName_Tags *map[string][]string, fkeys *[]string, res *map[string]string) {
	defaultt := ""
	(*res)[fName] = "INTEGER NOT NULL CHECK (" + fName + " IN (0, 1))"
	tags := (*mFieldName_Tags)[fName]
	for _, tag := range tags {
		if strings.Contains(tag, ":") {
			sp := strings.Split(tag, ":")
			switch sp[0] {
			case "fk":
				ref := strings.Split(sp[1], ".")
				if len(ref) == 2 {
					fkey := "FOREIGN KEY(" + fName + ") REFERENCES " + ref[0] + "(" + ref[1] + ")"
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
			}
		} else {
			logger.Error("tag", tag, "not handled for", fName, "of type", ty)
		}
		if defaultt != "" {
			(*res)[fName] += defaultt
		}
	}
}

func handleMigrationString(dialect, fName, ty string, mFieldName_Tags *map[string][]string, fkeys *[]string, res *map[string]string) {
	unique, notnull, text, defaultt, size, pk, check := "", "", "", "", "", "", ""
	tags := (*mFieldName_Tags)[fName]
	for _, tag := range tags {
		switch tag {
		case "unique":
			unique = " UNIQUE"
		case "text":
			text = " TEXT"
		case "notnull":
			notnull = " NOT NULL"
		case "pk":
			pk = " PRIMARY KEY"
		default:
			if strings.Contains(tag, ":") {
				sp := strings.Split(tag, ":")
				switch sp[0] {
				case "default":
					if sp[1] != "" {
						defaultt = " DEFAULT " + sp[1]
					} else {
						defaultt = " DEFAULT ''" 
					}
				case "fk":
					ref := strings.Split(sp[1], ".")
					if len(ref) == 2 {
						fkey := "FOREIGN KEY(" + fName + ") REFERENCES " + ref[0] + "(" + ref[1] + ")"
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
					sp := strings.Split(tag, ":")
					if sp[0] == "size" {
						size = sp[1]
					}
				case "check":
					if strings.Contains(strings.ToLower(sp[1]), "len") {
						switch dialect {
						case "sqlite", "":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "length", 1)
						case "postgres", "mysql":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "char_length", 1)
						default:
							logger.Error("check not handled for dialect:", dialect)
						}
					}
					check = " CHECK (" + sp[1] + ")"
				default:
					logger.Error("not handled", sp[0], "for", tag, ",field:", fName)
				}
			} else {
				logger.Error("tag", tag, "not handled for", fName, "of type", ty)
			}
		}
	}

	if text != "" {
		(*res)[fName] = text
	} else {
		if size != "" {
			(*res)[fName] = "VARCHAR(" + size + ")"
		} else {
			(*res)[fName] = "VARCHAR(255)"
		}
	}

	if unique != "" && pk == "" {
		(*res)[fName] += unique
	}
	if notnull != "" && pk == "" {
		(*res)[fName] += notnull
	}
	if pk != "" {
		(*res)[fName] += pk
	}
	if defaultt != "" {
		(*res)[fName] += defaultt
	}
	if check != "" {
		(*res)[fName] += check
	}
}

func handleMigrationFloat(dialect, fName, _ string, mFieldName_Tags *map[string][]string, fkeys *[]string, res *map[string]string) {
	mtags := map[string]string{}
	tags := (*mFieldName_Tags)[fName]
	for _, tag := range tags {
		switch tag {
		case "notnull":
			mtags["notnull"] = " NOT NULL"
		case "unique":
			mtags["unique"] = " UNIQUE"
		case "pk":
			mtags["pk"] = " PRIMARY KEY"
		default:
			if strings.Contains(tag, ":") {
				sp := strings.Split(tag, ":")
				switch sp[0] {
				case "default":
					if sp[1] != "" {
						mtags["default"] = " DEFAULT " + sp[1]
					}
				case "fk":
					ref := strings.Split(sp[1], ".")
					if len(ref) == 2 {
						fkey := "FOREIGN KEY(" + fName + ") REFERENCES " + ref[0] + "(" + ref[1] + ")"
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
				case "check":
					if strings.Contains(strings.ToLower(sp[1]), "len") {
						switch dialect {
						case "sqlite", "":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "length", 1)
						case "postgres", "mysql":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "char_length", 1)
						default:
							logger.Error("check not handled for dialect:", dialect)
						}
					}
					mtags["check"] = " CHECK (" + sp[1] + ")"
				default:
					logger.Error("not handled", sp[0], "for", tag, ",field:", fName)
				}
			}
		}

		(*res)[fName] = "DECIMAL(5,2)"
		for k, v := range mtags {
			switch k {
			case "pk":
				(*res)[fName] += v
			case "notnull":
				if _, ok := mtags["pk"]; !ok {
					(*res)[fName] += v
				}
			case "unique":
				if _, ok := mtags["pk"]; !ok {
					(*res)[fName] += v
				}
			case "default":
				(*res)[fName] += v
			case "check":
				(*res)[fName] += v
			default:
				logger.Error("case", k, "not handled")
			}
		}
	}
}

func handleMigrationTime(dialect, fName, ty string, mFieldName_Tags *map[string][]string, fkeys *[]string, res *map[string]string) {
	defaultt, notnull, check := "", "", ""
	tags := (*mFieldName_Tags)[fName]
	for _, tag := range tags {
		switch tag {
		case "now":
			switch dialect {
			case "sqlite", "":
				defaultt = "TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP"
			case "postgres":
				defaultt = "TIMESTAMP NOT NULL DEFAULT (now())"
			case "mysql":
				defaultt = "TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"
			default:
				logger.Error("not handled Time for ", fName, ty)
			}
		case "notnull":
			if defaultt != "" {
				notnull = " NOT NULL"
			}
		default:
			if strings.Contains(tag, ":") {
				sp := strings.Split(tag, ":")
				switch sp[0] {
				case "fk":
					ref := strings.Split(sp[1], ".")
					if len(ref) == 2 {
						fkey := "FOREIGN KEY(" + fName + ") REFERENCES " + ref[0] + "(" + ref[1] + ")"
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
				case "check":
					if strings.Contains(strings.ToLower(sp[1]), "len") {
						switch dialect {
						case "sqlite", "":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "length", 1)
						case "postgres", "mysql":
							sp[1] = strings.Replace(strings.ToLower(sp[1]), "len", "char_length", 1)
						default:
							logger.Error("check not handled for dialect:", dialect)
						}
					}
					check = " CHECK (" + sp[1] + ")"
				case "default":
					if sp[1] != "" {
						switch dialect {
						case "sqlite", "":
							defaultt = "TEXT NOT NULL DEFAULT " + sp[1]
						case "postgres":
							defaultt = "TIMESTAMP with time zone NOT NULL DEFAULT " + sp[1]
						case "mysql":
							defaultt = "TIMESTAMP with time zone NOT NULL DEFAULT " + sp[1]
						default:
							logger.Error("default for field", fName, "not handled")
						}
					}
				default:
					logger.Error("case", sp[0], "not handled")
				}
			}
		}
	}
	if defaultt != "" {
		(*res)[fName] = defaultt
	} else {
		if dialect == "" || dialect == "sqlite" {
			(*res)[fName] = "TEXT"
		} else {
			(*res)[fName] = "TIMESTAMP with time zone"
		}

		if notnull != "" {
			(*res)[fName] += notnull
		}
		if check != "" {
			(*res)[fName] += check
		}
	}
}

func prepareCreateStatement(tbName string, fields map[string]string, fkeys, cols []string,db *DatabaseEntity,ftags map[string][]string) string {
	st := "CREATE TABLE IF NOT EXISTS "
	st += tbName + " ("
	for i, col := range cols {
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
	for i, k := range fkeys {
		st += k
		if i < len(fkeys)-1 {
			st += ","
		}
	}
	st += ");"
	return st
}
