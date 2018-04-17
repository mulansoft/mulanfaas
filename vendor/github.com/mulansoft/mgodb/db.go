package mgodb

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "github.com/mulansoft/mgodb/utils"
)

type Database struct {
	session *mgo.Session
	latch   chan *mgo.Session
}

func (db *Database) Init(addr string, concurrent int, timeout time.Duration) {
	// create latch
	db.latch = make(chan *mgo.Session, concurrent)
	sess, err := mgo.Dial(addr)
	if err != nil {
		log.Println("mongodb: cannot connect to - ", addr, err)
		os.Exit(-1)
	}

	// set params
	sess.SetMode(mgo.Strong, true)
	sess.SetSocketTimeout(timeout)
	sess.SetCursorTimeout(0)
	db.session = sess

	for k := 0; k < cap(db.latch); k++ {
		db.latch <- sess.Copy()
	}
}

func (db *Database) Execute(f func(sess *mgo.Session) error) error {
	// latch control
	sess := <-db.latch
	defer func() {
		db.latch <- sess
	}()
	sess.Refresh()
	return f(sess)
}

var (
	_db Database
)

func Init(mongodb string, concurrent int, timeout time.Duration) {
	_db.Init(mongodb, concurrent, timeout)
}

func Execute(f func(sess *mgo.Session) error) error {
	return _db.Execute(f)
}

var (
	ErrModelNotPtr         = errors.New("model is not pointer")
	ErrModelToPtr          = errors.New("model point to another pointer")
	ErrCollectionNameIsNil = errors.New("model doesn't has collection name")
	ErrResultNotSliceAddr  = errors.New("result argument must be a slice address")
	ErrOperateFail         = errors.New("database operate fail")
)

// insert one record
// for example:
// user := &User{UserId: 1, Name: "xx"}
// Insert(user)
func Insert(model interface{}) error {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model": model,
			"err":   err,
		}).Error("insert db error: model validate fail")
		return err
	}

	updatedField := reflect.ValueOf(model).Elem().FieldByName("Updated")
	if updatedField.CanSet() {
		updatedField.Set(reflect.ValueOf(time.Now().UTC()))
	}
	createdField := reflect.ValueOf(model).Elem().FieldByName("Created")
	if createdField.CanSet() {
		createdField.Set(reflect.ValueOf(time.Now().UTC()))
	}

	collection := getCollectionName(model)
	err := Execute(func(sess *mgo.Session) error {
		return sess.DB("").C(collection).Insert(model)
	})
	if err != nil {
		log.WithFields(log.Fields{
			"model":      model,
			"collection": collection,
			"err":        err,
		}).Error("insert db error: database operate fail")
		return err
	}

	return err
}

// find one record
// for example:
// user := &User{}
// FindOne(user, bson.M{"name": "xxx"})
func FindOne(model interface{}, query interface{}) error {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model": model,
			"query": query,
			"err":   err,
		}).Error("find db error: model validate fail")
		return err
	}

	collection := getCollectionName(model)
	err := Execute(func(sess *mgo.Session) error {
		return sess.DB("").C(collection).Find(query).One(model)
	})
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"model":      model,
			"query":      query,
			"collection": collection,
			"err":        err,
		}).Error("find db error: database operate fail")
		return err
	}

	return err
}

// update one record
// for example
// user := &User{}
// UpdateOne(user, bson.M{"name": "xx"}, bson.M{"$set": bson.M{...}})
func UpdateOne(model interface{}, selector interface{}, update interface{}) error {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model":    model,
			"selector": selector,
			"update":   update,
			"err":      err,
		}).Error("update db error: validate model fail")
		return err
	}

	updatedField := reflect.ValueOf(model).Elem().FieldByName("Updated")
	if updatedField.CanSet() {
		updatedField.Set(reflect.ValueOf(time.Now().UTC()))
	}

	collection := getCollectionName(model)
	err := Execute(func(sess *mgo.Session) error {
		return sess.DB("").C(collection).Update(selector, update)
	})
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"model":      model,
			"selector":   selector,
			"update":     update,
			"collection": collection,
			"err":        err,
		}).Error("update db error: database operate fail")
	}

	return err
}

// upsert one record
// for example
// user := &User{"name":"xxx", "pwd": "xx"}
// user.UserId = 1
// UpsertOne(user, bson.M{"name": "xx"})
func UpsertOne(model interface{}, selector interface{}) error {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model":    model,
			"selector": selector,
			"err":      err,
		}).Error("upsert db error: validate model fail")
		return err
	}

	update := bson.M{"$set": model}
	err := UpdateOne(model, selector, update)
	if err == mgo.ErrNotFound {
		err = Insert(model)
	}
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"model":    model,
			"selector": selector,
			"err":      err,
		}).Error("upsert db error: database operate fail")
	}

	return err
}

// remove one record
// for example:
// user := &User{}
// RemoveOne(user, bson.M{"name": "xx"})
func RemoveOne(model interface{}, selector interface{}) error {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model":    model,
			"selector": selector,
			"err":      err,
		}).Error("delete db error: validate model fail")
		return err
	}

	collection := getCollectionName(model)
	err := Execute(func(sess *mgo.Session) error {
		return sess.DB("").C(collection).Remove(selector)
	})
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"model":      model,
			"selector":   selector,
			"collection": collection,
			"err":        err,
		}).Error("delete db error: database operate fail")
	}

	return err
}

// remove all record
// for example:
// user := &User{}
// RemoveAll(user, bson.M{"name": "xx"})
func RemoveAll(model interface{}, selector interface{}) error {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model":    model,
			"selector": selector,
			"err":      err,
		}).Error("delete all db error: validate model fail")
		return err
	}

	collection := getCollectionName(model)
	err := Execute(func(sess *mgo.Session) error {
		_, err := sess.DB("").C(collection).RemoveAll(selector)
		return err
	})
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"model":      model,
			"selector":   selector,
			"collection": collection,
			"err":        err,
		}).Error("delete all db error: database operate fail")
	}

	return err
}

// for example:
// result := []*User{}
// Find(&result, bson.M{...}, 1, 15, []string{...})
func Find(result interface{}, query interface{}, page int, pageSize int, sorts []string) error {
	if err := validateSlice(result); err != nil {
		log.WithFields(log.Fields{
			"result": result,
			"query":  query,
			"err":    err,
		}).Error("search db error: validate model fail")
		return err
	}

	collection := getCollectionName(result)
	skip := (page - 1) * pageSize
	err := Execute(func(sess *mgo.Session) error {
		if page < 0 && pageSize < 0 {
			return sess.DB("").C(collection).Find(query).Sort(sorts...).All(result)
		} else {
			return sess.DB("").C(collection).Find(query).Skip(skip).Limit(pageSize).Sort(sorts...).All(result)
		}
	})
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"result":   result,
			"query":    query,
			"page":     page,
			"pageSize": pageSize,
			"sorts":    sorts,
			"err":      err,
		}).Error("search db error: database operate fail")
	}

	return err
}

// for example:
// user := &User{}
// Count(user, bson.M{...})
func Count(model interface{}, query interface{}) int {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model": model,
			"query": query,
			"err":   err,
		}).Error("count db error: validate model fail")
		return 0
	}

	count := 0
	collection := getCollectionName(model)
	err := Execute(func(sess *mgo.Session) (err error) {
		count, err = sess.DB("").C(collection).Find(query).Count()
		return err
	})
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"model":      model,
			"query":      query,
			"collection": collection,
			"err":        err,
		}).Error("count db error: database operate fail")
		return 0
	}

	return count
}

// for example:
// user := &User{}
// UpdateAll(user, bson.M{...}, bson.M{"$set": bson.M{...}})
func UpdateAll(model interface{}, selector interface{}, update interface{}) (int, error) {
	if err := validateModel(model); err != nil {
		log.WithFields(log.Fields{
			"model":    model,
			"selector": selector,
			"update":   update,
			"err":      err,
		}).Error("update all db error: validate model fail")
		return 0, err
	}

	updatedField := reflect.ValueOf(model).Elem().FieldByName("Updated")
	if updatedField.CanSet() {
		updatedField.Set(reflect.ValueOf(time.Now().UTC()))
	}

	count := 0
	collection := getCollectionName(model)
	err := Execute(func(sess *mgo.Session) error {
		info, err := sess.DB("").C(collection).UpdateAll(selector, update)
		if !IsNil(info) {
			count = info.Updated
		}
		return err
	})
	if err != nil && err != mgo.ErrNotFound {
		log.WithFields(log.Fields{
			"model":      model,
			"selector":   selector,
			"update":     update,
			"collection": collection,
			"err":        err,
		}).Error("update all db error: database operate fail")
		return 0, err
	}

	return count, err
}

func validateModel(model interface{}) error {
	val := reflect.ValueOf(model)
	typ := reflect.Indirect(val).Type()

	// user := &User{"Name": "xx"}
	// InsertOneToDB(user, "UserId")
	if val.Kind() != reflect.Ptr {
		return ErrModelNotPtr
	}

	// For this case:
	// user := &User{"Name": "xx"}
	// InsertOneToDB(&user, "UserId")
	if typ.Kind() == reflect.Ptr {
		return ErrModelToPtr
	}

	// model must implement CollectionName() function
	if fun := val.MethodByName("CollectionName"); !fun.IsValid() {
		return ErrCollectionNameIsNil
	}

	return nil
}

func validateSlice(result interface{}) error {
	// result := []*User{}
	// SearchFromDB(&result, ...)
	val := reflect.ValueOf(result)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return ErrResultNotSliceAddr
	}

	// model must implement CollectionName() function
	val = reflect.Indirect(val)
	typ := val.Type().Elem()
	modelVal := reflect.New(typ)
	modelVal = reflect.Indirect(modelVal)
	if fun := modelVal.MethodByName("CollectionName"); !fun.IsValid() {
		return ErrCollectionNameIsNil
	}

	return nil
}

// 获取数据表名称
func getCollectionName(obj interface{}) string {
	var modelVal reflect.Value
	val := reflect.ValueOf(obj)
	if val.Elem().Kind() == reflect.Slice {
		val = reflect.Indirect(val)
		typ := val.Type().Elem()
		modelVal = reflect.New(typ)
		modelVal = reflect.Indirect(modelVal)
	} else {
		modelVal = val
	}

	if fun := modelVal.MethodByName("CollectionName"); fun.IsValid() {
		vals := fun.Call([]reflect.Value{})
		if len(vals) > 0 && vals[0].Kind() == reflect.String {
			return vals[0].String()
		}
	}

	// 如果没有定义表名，默认使用model类型名作表名
	// 比如ClubMsg，默认表名为club_msg
	return snakeString(reflect.Indirect(val).Type().Name())
}

// snake string, XxYy to xx_yy , XxYY to xx_yy
func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}
