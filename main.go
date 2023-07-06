package can2mqtt_json_legacy

import (
	//"bufio"        // Reader
	//"encoding/csv" // CSV Management
	//"encoding/binary"
	"encoding/json"
	"fmt" // print :)
	"io/ioutil"

	//"io"           // EOF const
	"log"     // error management
	"os"      // open files
	"strconv" // parse strings
	"strings"
	"sync"
)

// Generic CAN Frame

type Frame struct {
	// bit 0-28: CAN identifier (11/29 bit)
	// bit 29: error message flag (ERR)
	// bit 30: remote transmision request (RTR)
	// bit 31: extended frame format (EFF)
	ID     uint32
	Length uint8
	Flags  uint8
	Res0   uint8
	Res1   uint8
	Data   [8]uint8 // !!! hier für CAN FD -> max Framelength angeben  TODO
}

// Conversion Method
// Key identefies the data in json
// Type is for conversion Method
// Place says where the Data is inside the 8 byte array of canframe
// Factor is for calibrating ... maybe a offset is needet aswell TODO

type PayloadField struct {
	Key    string  `json:"key"`
	Type   string  `json:"type"`
	Place  [2]int  `json:"place"`
	Factor float64 `json:"factor"`
}

// stores all conversion for a frame

type Payload struct {
	Fields []PayloadField `json:"payload"`
}

// All nessesary data for Frame

type Conversion struct {
	Topic   string         `json:"topic"`
	CanID   string         `json:"canid"`
	Length  int            `json:"length"`
	Payload []PayloadField `json:"payload"`
}

// list for the two directions

type Config struct {
	Can2mqtt []Conversion `json:"can2mqtt"`
	Mqtt2can []Conversion `json:"mqtt2can"`
}

// whast going out from mqtt2can function

type mqtt_response struct {
	Topic   string
	Payload string
}

var config Config     // all config inside
var last_clock string=`00` // the last timestamp from dataquerry

// var pairFromID map[int]*can2mqtt       // c2m pair (lookup from ID)
// var pairFromTopic map[string]*can2mqtt // c2m pair (lookup from Topic)
var dbg = false                 // verbose on off [-v]
var ci = "can0"                 // the CAN-Interface [-c]
var cs = "tcp://localhost:1883" // mqtt-connect-string [-m]
var c2mf = "can2mqtt.json"      // path to the can2mqtt.json [-f]
var dirMode = 0                 // directional modes: 0=bidirectional 1=can2mqtt only 2=mqtt2can only [-d]
var wg sync.WaitGroup

// SetDbg decides whether there is really verbose output or
// just standard information output. Default is false.
func SetDbg(v bool) {
	dbg = v
}

// SetCi sets the CAN-Interface to use for the CAN side
// of the bridge. Default is: can0.
func SetCi(c string) {
	ci = c
}

// SetC2mf expects a string which is a path to a can2mqtt.csv file
// Default is: can2mqtt.csv
func SetC2mf(f string) {
	c2mf = f
}

// SetCs sets the MQTT connect-string which contains: protocol,
// hostname and port. Default is: tcp://localhost:1883
func SetCs(s string) {
	cs = s
}

// SetConfDirMode sets the dirMode
func SetConfDirMode(s string) {
	if s == "0" {
		dirMode = 0
	} else if s == "1" {
		dirMode = 1
	} else if s == "2" {
		dirMode = 2
	} else {
		_ = fmt.Errorf("error: got invalid value for -d (%s). Valid values are 0 (bidirectional), 1 (can2mqtt only) or 2 (mqtt2can only)", s)
	}
}

// Start is the function that should be called after debug-level
// connect-string, can interface and can2mqtt file have been set.
// Start takes care of everything that happens after that.
// It starts the CAN-Bus connection and the MQTT-Connection. It
// parses the can2mqtt.csv file and from there everything takes
// its course...
func Start() {
	fmt.Println("Starting can2mqtt")
	fmt.Println()
	fmt.Println("MQTT-Config:  ", cs)
	fmt.Println("CAN-Config:   ", ci)
	fmt.Println("can2mqtt.csv: ", c2mf)
	fmt.Print("dirMode:       ", dirMode, " (")
	if dirMode == 0 {
		fmt.Println("bidirectional)")
	}
	if dirMode == 1 {
		fmt.Println("can2mqtt only)")
	}
	if dirMode == 2 {
		fmt.Println("mqtt2can only)")
	}
	fmt.Print("Debug-Mode:    ")
	if dbg {
		fmt.Println("yes")
	} else {
		fmt.Println("no")
	}
	fmt.Println()
	wg.Add(1)
	go canStart(ci) // epic parallel shit ;-)
	mqttStart(cs)
	readC2MPFromFile(c2mf)
	wg.Wait()
}

// this functions opens, parses and extracts information out
// of the can2mqtt.csv
func readC2MPFromFile(filename string) {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Parsing Json

	// Decode the JSON data into a struct
	byteValue, _ := ioutil.ReadAll(file)

	// Write json as config
	json.Unmarshal(byteValue, &config)

	// subscribing to all MQTT topics

	for _, topic_tmp := range config.Mqtt2can {
		mqttSubscribe(topic_tmp.Topic)
	}

	for _, canid_tmp := range config.Can2mqtt {
		hexStr := canid_tmp.CanID
		if strings.HasPrefix(hexStr, "0x") {
			hexStr = strings.TrimPrefix(hexStr, "0x")
		}

		i, err := strconv.ParseUint(hexStr, 16, 32)
		if err != nil {
			fmt.Println(err)
			return
		}
		canSubscribe(uint32(i))
		if dbg {
			fmt.Printf("subscribed to :%x \n", i)
		}
	}

	/*
		//r := csv.NewReader(bufio.NewReader(file))
		//pairFromID = make(map[int]*can2mqtt)
		//pairFromTopic = make(map[string]*can2mqtt)
		for {
			//record, err := r.Read()
			// Stop at EOF.
			//if err == io.EOF {
			//	break
			//}
			//canID, err := strconv.Atoi(record[0])
			convMode := record[1]
			topic := record[2]
			if isInSlice(canID, topic) {
				panic("main: each ID and each topic is only allowed once!")
			}
			pairFromID[canID] = &can2mqtt{
				canId:      canID,
				convMethod: convMode,
				mqttTopic:  topic,
			}
			pairFromTopic[topic] = pairFromID[canID]
			mqttSubscribe(topic)        // TODO move to append function
			canSubscribe(uint32(canID)) // TODO move to append function
		}
	*/
	if dbg {
		fmt.Printf("main: the following CAN-MQTT pairs have been extracted:\n")
		fmt.Printf("Config for can2mqtt\n")
		for _, msg := range config.Can2mqtt {
			fmt.Println(msg.CanID, "\t\t", msg.Topic, "\t\t", msg.Payload)
		}
		fmt.Printf("Config for mqtt2can\n")
		for _, msg := range config.Mqtt2can {
			fmt.Println(msg.CanID, "\t\t", msg.Topic, "\t\t", msg.Payload)
		}
	}

}

/*
// check function to check if a topic or an ID is in the slice
func isInSlice(canId int, mqttTopic string) bool {
	if pairFromID[canId] != nil {
		if dbg {
			fmt.Printf("main: The ID %d or the Topic %s is already in the list!\n", canId, mqttTopic)
		}
		return true
	}
	if pairFromTopic[mqttTopic] != nil {
		if dbg {
			fmt.Printf("main: The ID %d or the Topic %s is already in the list!\n", canId, mqttTopic)
		}
		return true
	}
	return false
}

// get the corresponding ID for a given topic
func getIdFromTopic(topic string) int {
	return pairFromTopic[topic].canId
}

// get the conversion mode for a given topic
func getConvModeFromTopic(topic string) string {
	return pairFromTopic[topic].convMethod
}

// get the convertMode for a given ID
func getConvModeFromId(canId int) string {
	return pairFromID[canId].convMethod
}

// get the corresponding topic for an ID
func getTopicFromId(canId int) string {
	return pairFromID[canId].mqttTopic
}
*/