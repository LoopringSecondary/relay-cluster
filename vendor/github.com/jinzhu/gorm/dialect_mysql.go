package gorm

import (
	"crypto/sha1"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type mysql struct {
	commonDialect
}

func init() {
	RegisterDialect("mysql", &mysql{})
}

func (mysql) GetName() string {
	return "mysql"
}

func (mysql) Quote(key string) string {
	return fmt.Sprintf("`%s`", key)
}

// Get Data Type for MySQL Dialect
func (s *mysql) DataTypeOf(field *StructField) string {
	var dataValue, sqlType, size, additionalType = ParseFieldStructForDialect(field, s)

	// MySQL allows only one auto increment column per table, and it must
	// be a KEY column.
	if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok {
		if _, ok = field.TagSettings["INDEX"]; !ok && !field.IsPrimaryKey {
			delete(field.TagSettings, "AUTO_INCREMENT")
		}
	}

	if sqlType == "" {
		switch dataValue.Kind() {
		case reflect.Bool:
			sqlType = "boolean"
		case reflect.Int8:
<<<<<<< HEAD
			if s.fieldCanAutoIncrement(field) {
=======
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok || field.IsPrimaryKey {
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "tinyint AUTO_INCREMENT"
			} else {
				sqlType = "tinyint"
			}
		case reflect.Int, reflect.Int16, reflect.Int32:
<<<<<<< HEAD
			if s.fieldCanAutoIncrement(field) {
=======
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok || field.IsPrimaryKey {
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "int AUTO_INCREMENT"
			} else {
				sqlType = "int"
			}
		case reflect.Uint8:
<<<<<<< HEAD
			if s.fieldCanAutoIncrement(field) {
=======
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok || field.IsPrimaryKey {
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "tinyint unsigned AUTO_INCREMENT"
			} else {
				sqlType = "tinyint unsigned"
			}
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
<<<<<<< HEAD
			if s.fieldCanAutoIncrement(field) {
=======
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok || field.IsPrimaryKey {
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "int unsigned AUTO_INCREMENT"
			} else {
				sqlType = "int unsigned"
			}
		case reflect.Int64:
<<<<<<< HEAD
			if s.fieldCanAutoIncrement(field) {
=======
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok || field.IsPrimaryKey {
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "bigint AUTO_INCREMENT"
			} else {
				sqlType = "bigint"
			}
		case reflect.Uint64:
<<<<<<< HEAD
			if s.fieldCanAutoIncrement(field) {
=======
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok || field.IsPrimaryKey {
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
				field.TagSettings["AUTO_INCREMENT"] = "AUTO_INCREMENT"
				sqlType = "bigint unsigned AUTO_INCREMENT"
			} else {
				sqlType = "bigint unsigned"
			}
		case reflect.Float32, reflect.Float64:
			sqlType = "double"
		case reflect.String:
			if size > 0 && size < 65532 {
				sqlType = fmt.Sprintf("varchar(%d)", size)
			} else {
				sqlType = "longtext"
			}
		case reflect.Struct:
			if _, ok := dataValue.Interface().(time.Time); ok {
<<<<<<< HEAD
				precision := ""
				if p, ok := field.TagSettings["PRECISION"]; ok {
					precision = fmt.Sprintf("(%s)", p)
				}

				if _, ok := field.TagSettings["NOT NULL"]; ok {
					sqlType = fmt.Sprintf("timestamp%v", precision)
				} else {
					sqlType = fmt.Sprintf("timestamp%v NULL", precision)
=======
				if _, ok := field.TagSettings["NOT NULL"]; ok {
					sqlType = "timestamp"
				} else {
					sqlType = "timestamp NULL"
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
				}
			}
		default:
			if IsByteArrayOrSlice(dataValue) {
				if size > 0 && size < 65532 {
					sqlType = fmt.Sprintf("varbinary(%d)", size)
				} else {
					sqlType = "longblob"
				}
			}
		}
	}

	if sqlType == "" {
		panic(fmt.Sprintf("invalid sql type %s (%s) for mysql", dataValue.Type().Name(), dataValue.Kind().String()))
	}

	if strings.TrimSpace(additionalType) == "" {
		return sqlType
	}
	return fmt.Sprintf("%v %v", sqlType, additionalType)
}

func (s mysql) RemoveIndex(tableName string, indexName string) error {
	_, err := s.db.Exec(fmt.Sprintf("DROP INDEX %v ON %v", indexName, s.Quote(tableName)))
	return err
}

<<<<<<< HEAD
func (s mysql) ModifyColumn(tableName string, columnName string, typ string) error {
	_, err := s.db.Exec(fmt.Sprintf("ALTER TABLE %v MODIFY COLUMN %v %v", tableName, columnName, typ))
	return err
}

=======
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
func (s mysql) LimitAndOffsetSQL(limit, offset interface{}) (sql string) {
	if limit != nil {
		if parsedLimit, err := strconv.ParseInt(fmt.Sprint(limit), 0, 0); err == nil && parsedLimit >= 0 {
			sql += fmt.Sprintf(" LIMIT %d", parsedLimit)

			if offset != nil {
				if parsedOffset, err := strconv.ParseInt(fmt.Sprint(offset), 0, 0); err == nil && parsedOffset >= 0 {
					sql += fmt.Sprintf(" OFFSET %d", parsedOffset)
				}
			}
		}
	}
	return
}

func (s mysql) HasForeignKey(tableName string, foreignKeyName string) bool {
	var count int
<<<<<<< HEAD
	currentDatabase, tableName := currentDatabaseAndTable(&s, tableName)
	s.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=? AND TABLE_NAME=? AND CONSTRAINT_NAME=? AND CONSTRAINT_TYPE='FOREIGN KEY'", currentDatabase, tableName, foreignKeyName).Scan(&count)
=======
	s.db.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=? AND TABLE_NAME=? AND CONSTRAINT_NAME=? AND CONSTRAINT_TYPE='FOREIGN KEY'", s.CurrentDatabase(), tableName, foreignKeyName).Scan(&count)
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
	return count > 0
}

func (s mysql) CurrentDatabase() (name string) {
	s.db.QueryRow("SELECT DATABASE()").Scan(&name)
	return
}

func (mysql) SelectFromDummyTable() string {
	return "FROM DUAL"
}

<<<<<<< HEAD
func (s mysql) BuildKeyName(kind, tableName string, fields ...string) string {
	keyName := s.commonDialect.BuildKeyName(kind, tableName, fields...)
=======
func (s mysql) BuildForeignKeyName(tableName, field, dest string) string {
	keyName := s.commonDialect.BuildForeignKeyName(tableName, field, dest)
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
	if utf8.RuneCountInString(keyName) <= 64 {
		return keyName
	}
	h := sha1.New()
	h.Write([]byte(keyName))
	bs := h.Sum(nil)

<<<<<<< HEAD
	// sha1 is 40 characters, keep first 24 characters of destination
	destRunes := []rune(regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(fields[0], "_"))
=======
	// sha1 is 40 digits, keep first 24 characters of destination
	destRunes := []rune(regexp.MustCompile("(_*[^a-zA-Z]+_*|_+)").ReplaceAllString(dest, "_"))
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
	if len(destRunes) > 24 {
		destRunes = destRunes[:24]
	}

	return fmt.Sprintf("%s%x", string(destRunes), bs)
}
<<<<<<< HEAD

func (mysql) DefaultValueStr() string {
	return "VALUES()"
}
=======
>>>>>>> 258d5c409a01370dfe542ceadc3d1669659150fe
