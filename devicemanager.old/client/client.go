package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	deviceApi "github.com/michaelhenkel/dmng/devicemanager/client/api"
	dmPB "github.com/michaelhenkel/dmng/devicemanager/protos"
	"gopkg.in/yaml.v2"
)

var (
	request        = flag.String("request", "", "path to input file")
	interfaceMap   = make(map[string]*dmPB.Interface)
	connectionMap  = make(map[string]deviceApi.DMClient)
	connectionMap2 = make(map[string]*deviceApi.DMClient)
)

func jsonPrettyPrint(in []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, in, "", "\t")
	if err != nil {
		return string(in)
	}
	return out.String()
}

/*

func (d *Device) sendConfig(device string) error {
	var intfList []*dmPB.Interface
	if len(interfaceMap) > 0 {
		for _, intf := range interfaceMap {
			intfList = append(intfList, intf)
		}
		d.deviceConnection.CreateInterface(intfList)

			if err != nil {
				return err
			}

	}
	return nil
}

func (d *Device) devContext(device string, reader *bufio.Reader) error {
	for {
		fmt.Printf("%s -> ", device)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if strings.Compare("exit", text) == 0 {
			fmt.Println("Exiting")
			break
		}
		if strings.Compare("create", text) == 0 {
			fmt.Println("Entering Create context")
			if err := d.createContext(device, reader); err != nil {
				return err
			}
		}
		if strings.Compare("send", text) == 0 {
			fmt.Println("Sending config")
			if err := d.sendConfig(device); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Device) createContext(device string, reader *bufio.Reader) error {
	for {
		fmt.Printf("%s -> create -> ", device)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if strings.Compare("exit", text) == 0 {
			fmt.Println("Exiting")
			break
		}
		textSlice := strings.Split(text, " ")
		if strings.Compare("interface", textSlice[0]) == 0 {
			if len(textSlice) == 1 {
				fmt.Println("interace name must be specified")
			} else {
				if err := d.createInterface(textSlice[1:]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
*/

func (d *Device) createInterface(intf []string) error {
	intfName := intf[0]
	var ipv4 string
	var ipv6 string
	for idx, intfField := range intf[1:] {
		switch intfField {
		case "ipv4":
			ipv4 = intf[idx+2]
		case "ipv6":
			ipv6 = intf[idx+2]
		}
	}
	intfObj := &dmPB.Interface{
		Name: intfName,
	}
	if ipv4 != "" {
		intfObj.Ipv4 = ipv4
	}
	if ipv6 != "" {
		intfObj.Ipv6 = ipv6
	}
	interfaceMap[intfName] = intfObj
	return nil
}

type Requests struct {
	Request []map[string][]Input `yaml:"request"`
}

type Input struct {
	Op         string            `yaml:"op"`
	Interfaces []*dmPB.Interface `yaml:"interfaces"`
}

/*
	go func() {
		for i := 0; i < 20; i++ {
			msg := &pb.SimpleData{Msg: "msg " + strconv.Itoa(i)}
			rand.Seed(time.Now().UnixNano())
			n := rand.Intn(20)
			log.Printf("sleeping for %d seconds\n", n)
			log.Println("sending msg: " + msg.GetMsg())
			time.Sleep(time.Duration(n) * time.Second)
			run.Send <- msg
		}
	}()
	go func() {
		for {
			select {
			case msg := <-run.Receive:
				log.Println("received reply: " + msg.GetMsg())
			default:
			}
		}
	}()
		dmClient := &DMClient{
		Stream:  stream,
		Request: make(chan *pbDM.Request),
		Result:  make(chan *pbDM.Result),
	}
*/

func main() {
	flag.Parse()
	if *request != "" {
		rquestYaml, err := ioutil.ReadFile(*request)
		if err != nil {
			log.Fatalln(err)
		}
		var requestConfig Requests
		err = yaml.Unmarshal(rquestYaml, &requestConfig)
		if err != nil {
			fmt.Printf("Error parsing YAML file: %s\n", err)
		}
		for _, request := range requestConfig.Request {
			for server := range request {
				var serverConnection *deviceApi.DMClient
				var ok bool
				if serverConnection, ok = connectionMap2[server]; !ok {
					serverConnection = &deviceApi.DMClient{
						Address: server,
						Request: make(chan *dmPB.Request),
						Result:  make(chan *dmPB.Result),
					}
					connectionMap2[server] = serverConnection
				}
				go serverConnection.NewClient()
			}

			for server, inputs := range request {
				s := connectionMap2[server]
				done := make(chan bool)
				go func() {
					for _, input := range inputs {
						switch input.Op {
						case "create":
							intfList := &dmPB.Interfaces{
								Interface: input.Interfaces,
							}
							request := &dmPB.Request{
								Request: &dmPB.Request_Create{
									Create: &dmPB.Create{
										CreateRequest: &dmPB.Create_Interfaces{
											Interfaces: intfList,
										},
									},
								},
							}
							log.Println("sending request")
							//s.Request <- request
							s.SendRquest(request)
							fmt.Println("bla")
						}
					}
					//close(done)
				}()
				go func() {
					for {
						select {
						case result := <-s.Result:
							log.Println("got result ", result)
						default:
						}
					}

				}()
				<-done
			}
		}
	}
	/*
		else {
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("Simple Shell")
			fmt.Println("---------------------")

			for {
				fmt.Print("-> ")
				text, _ := reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)
				textSlice := strings.Split(text, " ")
				if strings.Compare("exit", text) == 0 {
					fmt.Println("Exiting")
					break
				}
				if strings.Compare("device", textSlice[0]) == 0 {
					if len(textSlice) == 1 {
						fmt.Println("device must be specified")
					} else {
						fmt.Println("Entering Device context")
						c := deviceApi.Connection{ServerAddress: textSlice[1]}
						if err := c.NewClient(); err != nil {
							log.Fatalln(err)
						}
						device := &Device{
							deviceConnection: &c,
						}
						device.devContext(textSlice[1], reader)
					}
				}
			}

		}
	*/
}

type Device struct {
	deviceConnection *deviceApi.DMClient
}
