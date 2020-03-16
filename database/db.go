package database

import (
	"io/ioutil"
	"log"
	"os"

	dmPB "github.com/michaelhenkel/dmng/devicemanager/protos"
	"gopkg.in/yaml.v2"
)

func Write() {

}

func Read() {

}

func List() {

}

func Update() {

}

func Delete() {

}

type dbStructure struct {
	Interfaces []dmPB.Interface
	Devices    []dmPB.Device
}

type DB interface {
	Write()
	Read()
	List()
	Update()
	Delete()
}

func (c *conf) readDB() *conf {
	dbFileString := "db.yaml"
	dbFile, err := ioutil.ReadFile(dbFileString)
	if err != nil && os.IsNotExist(err) {
		log.Printf("file doesn't exist, creatu   #%v ", err)
		emptyFile, err := os.Create(dbFileString)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(emptyFile)
		emptyFile.Close()
		dbFile, err = ioutil.ReadFile(dbFileString)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
