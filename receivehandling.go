package can2mqtt_json_legacy

import (
	"fmt"

	"github.com/brutella/can"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// handleCAN is the standard receive handler for CANFrames
// and does the following:
// 1. calling standard convert function: convert2MQTT
// 2. sending the message
func handleCAN(cf can.Frame) {
	if dbg {
		fmt.Printf("receivehandler: received CANFrame: ID: %d, len: %d, payload %s\n", cf.ID, cf.Length, cf.Data)
	}
	mqttPayload := convert2MQTT(int(cf.ID), int(cf.Length), cf.Data)
	if dbg {
		fmt.Printf("receivehandler: converted String: %s\n", mqttPayload)
	}

	if dirMode != 2 {
		mqttPublish(mqttPayload.Topic, mqttPayload.Payload)
		fmt.Printf("ID: %d len: %d data: %X -> topic: \"%s\" message: \"%s\"\n", cf.ID, cf.Length, cf.Data, mqttPayload.Topic, mqttPayload.Payload)
	}
}

// handleMQTT is the standard receive handler for MQTT
// messages and does the following:
// 1. calling the standard convert function: convert2CAN
// 2. sending the message
func handleMQTT(_ MQTT.Client, msg MQTT.Message) {
	if dbg {
		fmt.Printf("receivehandler: received message: topic: %s, msg: %s\n", msg.Topic(), msg.Payload())
	}
	cf := convert2CAN(msg.Topic(), string(msg.Payload()))

	if dirMode != 1 {
		canPublish(cf)
		fmt.Printf("ID: %x len: %d data: %X <- topic: \"%s\" message: \"%s\"\n", cf.ID, cf.Length, cf.Data, msg.Topic(), msg.Payload())
	}
}
