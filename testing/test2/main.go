package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

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
				devContext(textSlice[1], reader)
			}
		}

	}

}

func devContext(device string, reader *bufio.Reader) {
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
			createContext(device, reader)
		}
	}
}

func createContext(device string, reader *bufio.Reader) {
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
				createInterface(textSlice[1:])
			}
		}
	}
}

func createInterface(intf []string) {
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
	intfString := intfName
	if ipv4 != "" {
		intfString = intfString + " ipv4 " + ipv4
	}
	if ipv6 != "" {
		intfString = intfString + " ipv6 " + ipv6
	}
	fmt.Println(intfString)
}
