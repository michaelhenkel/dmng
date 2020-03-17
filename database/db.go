package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	dmPB "github.com/michaelhenkel/dmng/devicemanager/protos"
)

func (dbclient *db) ReadDevice(device *dmPB.Device) error {
	for _, obj := range dbclient.Devices {
		if device.GetName() == obj.GetName() {
			device = &obj
			return nil
		}
	}
	return &ObjectNotFound{}
}

func (dbclient *db) CreateDevice(device *dmPB.Device) error {
	var devices []dmPB.Device
	if dbclient.Devices == nil {
		dbclient.Devices = devices
	}
	dbclient.Devices = append(dbclient.Devices, *device)
	if err := writeDB(dbclient); err != nil {
		return err
	}
	return nil
}

func (dbclient *db) ReadInterface(obj *dmPB.Interface) error {
	for _, elem := range dbclient.Interfaces {
		if obj.GetName() == elem.GetName() {
			obj = &elem
			return nil
		}
	}
	return &ObjectNotFound{}
}

func (dbclient *db) CreateInterface(obj *dmPB.Interface) error {
	var interfaces []dmPB.Interface
	if dbclient.Interfaces == nil {
		dbclient.Interfaces = interfaces
	}
	dbclient.Interfaces = append(dbclient.Interfaces, *obj)
	if err := writeDB(dbclient); err != nil {
		return err
	}
	return nil
}

func (dbclient *db) DeleteInterface(obj *dmPB.Interface) error {
	if dbclient.Interfaces == nil {
		return fmt.Errorf("no interfaces")
	}
	var idxObj int
	for idx, elem := range dbclient.Interfaces {
		if obj.GetName() == elem.GetName() {
			idxObj = idx
			break
		}
	}
	dbclient.Interfaces[idxObj] = dbclient.Interfaces[len(dbclient.Interfaces)-1]
	dbclient.Interfaces[len(dbclient.Interfaces)-1] = dmPB.Interface{}
	dbclient.Interfaces = dbclient.Interfaces[:len(dbclient.Interfaces)-1]
	if err := writeDB(dbclient); err != nil {
		return err
	}
	return nil
}

type ObjectNotFound struct{}

func (f ObjectNotFound) Error() string {
	return fmt.Sprintln("object not found")
}

type ObjectAlreadyExists struct{}

func (f ObjectAlreadyExists) Error() string {
	return fmt.Sprintln("object already exists")
}

type db struct {
	Interfaces []dmPB.Interface
	Devices    []dmPB.Device
}

func NewDBClient() *db {
	return readDB()
}

type DBClient interface {
	ReadDevice(device *dmPB.Device) error
	CreateDevice(device *dmPB.Device) error
}

var (
	dbclient     *db
	dbFileString = "db.json"
)

func init() {
	dbclient = readDB()
}

func writeDB(db *db) error {
	dbFile, err := json.Marshal(db)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dbFileString, dbFile, os.FileMode(600))
	if err != nil {
		return err
	}
	return nil
}

func readDB() *db {
	var db db
	dbFile, err := ioutil.ReadFile(dbFileString)
	if err != nil && os.IsNotExist(err) {
		log.Printf("file doesn't exist, creating...\n")
		emptyFile, err := os.Create(dbFileString)
		if err != nil {
			log.Fatal(err)
		}
		emptyFile.Close()
		dbFile, err = ioutil.ReadFile(dbFileString)
	}
	if len(dbFile) > 0 {
		err = json.Unmarshal(dbFile, &db)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	}
	return &db
}
