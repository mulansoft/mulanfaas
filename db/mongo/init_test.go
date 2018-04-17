package mongo

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	db "github.com/mulansoft/mgodb"
)

const (
	CAR_COLLECTION       = "car"
	OWNER_COLLECTION     = "owner"
	CAR_OWNER_COLLECTION = "car_owner"
)

type Car struct {
	CarId   int64       `json:"carId" bson:"carId"`
	Name    string      `json:"name" bson:"name"`
	Price   int         `json:"price" bson:"price"`
	Remark  interface{} `json:"remark" bson:"remark"`
	Updated time.Time   `json:"updated" bson:"updated"`
	Created time.Time   `json:"created" bson:"Created"`
}

type Owner struct {
	OwnerId int64  `json:"ownerId" bson:"ownerId"`
	Name    string `json:"name" bson:"name"`
}

type CarOwner struct {
	OwnerId int64   `json:"ownerId" bson:"ownerId"`
	CarId   int64   `json:"carId" bson:"carId"`
	Cars    []Car   `bson:"cars,omitempty"`
	Owners  []Owner `bson:"owners,omitempty"`
}

func (c *Car) CollectionName() string {
	return CAR_COLLECTION
}

func (o *Owner) CollectionName() string {
	return OWNER_COLLECTION
}

func (co *CarOwner) CollectionName() string {
	return CAR_OWNER_COLLECTION
}

func NewCar() *Car {
	obj := &Car{CarId: getUUID()}
	return obj
}

func getUUID() int64 {
	return int64(rand.Uint32())
}

func TestBsonRemark(t *testing.T) {
	Init()

	// new car with remark
	car := NewCar()
	car.Name = "东风风行"
	car.Price = 80000
	car.Remark = bson.M{
		"remark1": car.CarId,
		"remark2": car.CarId,
	}
	db.Insert(car)

	// find car by remark
	obj := new(Car)
	if err := db.FindOne(obj, bson.M{"remark.remark1": car.CarId}); err != nil {
		t.Error("test bson remark error")
	} else {
		jsonData, _ := json.Marshal(obj)
		t.Log(string(jsonData))
	}
}

func TestAggregate(t *testing.T) {
	Init()
	// new car
	car := new(Car)
	car.CarId = 6330682874475319296
	car.Name = "本田思域"
	car.Price = 120000
	db.Insert(car)

	// new owner
	owner := new(Owner)
	owner.OwnerId = 6330682222932135936
	owner.Name = "Simi"
	db.Insert(owner)

	// car_owner
	co := new(CarOwner)
	co.CarId = car.CarId
	co.OwnerId = owner.OwnerId
	db.Insert(co)

	// aggregate
	pipeline := []bson.M{
		bson.M{"$match": bson.M{"ownerId": owner.OwnerId}},
		bson.M{
			"$lookup": bson.M{
				"from":         car.CollectionName(),
				"localField":   "carId",
				"foreignField": "carId",
				"as":           "cars",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         owner.CollectionName(),
				"localField":   "ownerId",
				"foreignField": "ownerId",
				"as":           "owners",
			},
		},
	}
	resp := []*CarOwner{}
	err := db.Execute(func(sess *mgo.Session) error {
		return sess.DB("").C(CAR_OWNER_COLLECTION).Pipe(pipeline).All(&resp)
	})
	fmt.Println(err)

	// print resp
	for _, item := range resp {
		fmt.Println(item.Cars[0].Name, item.Owners[0].Name)
	}
}

func TestCRUD(t *testing.T) {
	Init()

	car := NewCar()
	car.Name = "奔驰"
	car.Price = 100

	// 插入功能
	err := db.Insert(car)
	throwFail(t, err)
	if err != nil {
		throwFail(t, errors.New("id value not equal"))
	}
	t.Logf("insert result, id:%v, car:%v", car.CarId, car)

	// 搜索功能
	car1 := NewCar()
	err = db.FindOne(car1, bson.M{"carId": car.CarId})
	throwFail(t, err)
	t.Logf("find result: %v", car1)

	// 更新功能
	err = db.UpdateOne(car1, bson.M{"carId": car.CarId}, bson.M{"$set": bson.M{"name": "BMW"}})
	throwFail(t, err)
	car2 := NewCar()
	err = db.FindOne(car2, bson.M{"carId": car.CarId})
	throwFail(t, err)
	if car2.Name != "BMW" {
		throwFail(t, errors.New("UpdateOne fail"))
	}
	t.Logf("update result: %v", car2)

	// Upsert功能
	car4 := NewCar()
	car4.CarId = getUUID()
	car4.Name = "吉瑞QQ"
	car4.Price = 15
	err = db.UpsertOne(car4, bson.M{"carId": car4.CarId})
	throwFail(t, err)
	car5 := NewCar()
	err = db.FindOne(car5, bson.M{"carId": car4.CarId})
	throwFail(t, err)
	t.Logf("upsert result: %v", car5)

	// 分页功能
	result := []*Car{}
	err = db.Find(&result, bson.M{}, 1, 10, []string{"-created"})
	throwFail(t, err)
	t.Logf("search result: %v", result)

	// 删除功能
	car3 := NewCar()
	err = db.RemoveOne(car3, bson.M{"carId": car.CarId})
	throwFail(t, err)
	err = db.FindOne(car3, bson.M{"carId": car.CarId})
	if err != mgo.ErrNotFound {
		throwFail(t, err)
	}
}

func TestUpsertOne(t *testing.T) {
	Init()

	car := NewCar()
	car.CarId = getUUID()
	car.Name = "宝马X5"
	car.Price = 100
	err := db.UpsertOne(car, bson.M{"carId": car.CarId})
	throwFail(t, err)

	time.Sleep(5 * time.Second)

	car.Price = 150
	err = db.UpsertOne(car, bson.M{"carId": car.CarId})
	throwFail(t, err)
}

func throwFail(t *testing.T, err error) {
	if err != nil {
		info := fmt.Sprintf("\t\nError: %s", err.Error())
		t.Errorf(info)
		t.Fail()
	}
}
