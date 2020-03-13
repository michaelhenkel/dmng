package inventory

import (
	"io/ioutil"
	"log"

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

type DB interface {
	Write()
	Read()
	List()
	Update()
	Delete()
}

func (c *conf) getConf() *conf {

	dbFile, err := ioutil.ReadFile("db.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
