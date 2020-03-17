package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	deviceApi "github.com/michaelhenkel/dmng/devicemanager/client/api"
	dmPB "github.com/michaelhenkel/dmng/devicemanager/protos"
)

var (
	serverAddr = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	addport    = flag.String("addport", "", "port to be added")
	getport    = flag.String("getport", "", "port to be retrieved")
	delport    = flag.String("delport", "", "port to be deleted")
)

func jsonPrettyPrint(in []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, in, "", "\t")
	if err != nil {
		return string(in)
	}
	return out.String()
}

func main() {
	flag.Parse()
	deviceClient := &deviceApi.Connection{
		ServerAddress: *serverAddr,
	}
	if *addport != "" {
		var intfList []*dmPB.Interface
		intfSlice := strings.Split(*addport,",")
		for _, intf := range intfSlice{
		_intf := &dmPB.Interface{
			Name: intf,
		}
		intfList = append(intfList, _intf)
	}
		err := deviceClient.CreateInterface(intfList)
		if err != nil {
			log.Fatalln(err)
		}
		//log.Println(result.Msg)
	}
	if *getport != "" {
		intf := &dmPB.Interface{
			Name: *getport,
		}
		_, err := deviceClient.ReadInterface(intf)
		if err != nil {
			log.Fatalln(err)
		}
		intfJSON, err := json.Marshal(intf)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%s\n", jsonPrettyPrint(intfJSON))
	}

	if *delport != "" {
		intf := &dmPB.Interface{
			Name: *delport,
		}
		_, err := deviceClient.DeleteInterface(intf)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
