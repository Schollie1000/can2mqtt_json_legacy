package can2mqtt_json_legacy

import (
	"encoding/binary"
	"encoding/json"

	//"encoding/hex"
	"fmt"
	"strconv"

	"math"
	"strings"

	"github.com/brutella/can"
)

func getPayloadconv(config *Config, id string, mode string) (*Payload, string) {
	var tmode []Conversion
	var id_compare string
	if mode == "can2mqtt" {
		tmode = config.Can2mqtt
		//id_compare = conversion.CanID
	} else if mode == "mqtt2can" {
		tmode = config.Mqtt2can
		//id_compare = conversion.Topic
	} else {
		return nil, ""
	}

	for _, conversion := range tmode {
		if mode == "can2mqtt" {
			id_compare = conversion.CanID
		} else {
			id_compare = conversion.Topic
		}
		//fmt.Println(id_compare)
		if id_compare == id {
			//fmt.Println("Found matching conversion in Can2mqtt")
			//fmt.Println("Conversion: ", conversion)      // Debug print
			//fmt.Println("Payload: ", conversion.Payload) // Debug print
			payload := Payload{}

			for _, field := range conversion.Payload {
				payloadField := PayloadField{
					Key:    field.Key,
					Type:   field.Type,
					Place:  field.Place,
					Factor: field.Factor,
				}

				payload.Fields = append(payload.Fields, payloadField)
			}
			if mode == "can2mqtt" {
				return &payload, conversion.Topic
			} else {
				return &payload, conversion.CanID
			}
		}
	}
	fmt.Println("No matching conversion found in Can2mqtt")
	errorPay := Payload{}
	errorField := PayloadField{
		Key:    "error",
		Type:   "error",
		Place:  [2]int{0, 0},
		Factor: 0,
	}
	errorPay.Fields = append(errorPay.Fields, errorField)
	return &errorPay, ""
}
func convert2MQTT(id int, length int, payload [8]byte) mqtt_response {
	idStr := fmt.Sprintf("0x%X", id)
	fmt.Printf("id = %s\n", idStr)
	conv, topic := getPayloadconv(&config, idStr, "can2mqtt")
	retstr := strings.Builder{}
	retstr.WriteString("{")
	var valstring string
	for _, field := range conv.Fields {
		valstring = ""
		switch field.Type {
		case "error":
			valstring = "error"
		case "unixtime":
			unix := binary.LittleEndian.Uint32(payload[0:4])
			ms := binary.LittleEndian.Uint32(payload[4:8])
			unixf := float64(unix)
			msf := float64(ms) / 1000
			valstring = strconv.FormatFloat(float64(unixf+msf), 'g', -1, 64)
			last_clock = valstring
		case "byte":
			sub := payload[field.Place[0]:field.Place[1]]
			if dbg {
				fmt.Printf("byte detected ")
				fmt.Printf(" sub  %x \n", sub)
			}
			valstring = string(sub)
		case "int8_t":
			sub := payload[field.Place[0]]
			if dbg {
				fmt.Printf("int 8 detected ")
				fmt.Printf(" sub  %x \n", sub)
			}
			data2 := int8(sub)
			tmpf := field.Factor * float64(data2)
			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
		case "uint8_t":
			sub := payload[field.Place[0]]
			if dbg {
				fmt.Printf("uint 8 detected ")
				fmt.Printf(" sub  %x \n", sub)
			}
			data2 := sub
			tmpf := field.Factor * float64(data2)
			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
		case "int16_t":
			sub := payload[field.Place[0]:field.Place[1]]
			if dbg {
				fmt.Printf("int 16 detected ")
				fmt.Printf(" sub  %x %x \n", sub[0], sub[1])
			}
			data2 := int16(binary.LittleEndian.Uint16(sub))
			tmpf := field.Factor * float64(data2)
			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
		case "uint16_t":
			sub := payload[field.Place[0]:field.Place[1]]
			if dbg {
				fmt.Printf("uint 16 detected ")
				fmt.Printf(" sub  %x %x \n", sub[0], sub[1])
			}
			data2 := binary.LittleEndian.Uint16(sub)
			tmpf := field.Factor * float64(data2)
			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
		case "int32_t":
			sub := payload[field.Place[0]:field.Place[1]]
			if dbg {
				fmt.Printf("int 32 detected ")
				fmt.Printf(" sub  %x %x %x %x\n", sub[0], sub[1], sub[2], sub[3])
			}
			data2 := int32(binary.LittleEndian.Uint32(sub))
			tmpf := field.Factor * float64(data2)
			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
		case "uint32_t":
			sub := payload[field.Place[0]:field.Place[1]]
			if dbg {
				fmt.Printf("uint 32 detected ")
				fmt.Printf(" sub  %x %x %x %x\n", sub[0], sub[1], sub[2], sub[3])
			}
			data2 := binary.LittleEndian.Uint32(sub)
			tmpf := field.Factor * float64(data2)
			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
		case "float":
			sub := payload[field.Place[0]:field.Place[1]]
			if dbg {
				fmt.Printf("float 32 detected ")
				fmt.Printf(" sub  %x %x %x %x\n", sub[0], sub[1], sub[2], sub[3])
			}
			data3 := binary.LittleEndian.Uint32(sub)
			data2 := math.Float32frombits(data3)
			tmpf := field.Factor * float64(data2)
			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
		}
		retstr.WriteString(`"` + field.Key + `" : ` + valstring + ", ")
	}
	if topic != "clock" {
		retstr.WriteString(`"unixtime" : ` + last_clock)
	}
	finalStr := retstr.String()
	finalStr = strings.TrimSuffix(finalStr, ", ")
	finalStr += "}"
	res := mqtt_response{}
	res.Topic = topic
	res.Payload = finalStr
	return res
}

// func convert2MQTT(id int, length int, payload [8]byte) mqtt_response {
// 	idStr := fmt.Sprintf("0x%X", id)
// 	fmt.Printf("id = %s\n", idStr)
// 	conv, topic := getPayloadconv(&config, idStr, "can2mqtt")
// 	retstr := "{"	
// 	var valstring string 
// 	for _, field := range conv.Fields {
// 		valstring = ""
// 		if field.Type == "error" {
// 			valstring = "error"
// 		} else if field.Type == "unixtime" {
			
// 			unix := uint32(payload[0]) | uint32(payload[1])<<8 | uint32(payload[2])<<16 | uint32(payload[3])<<24
// 			ms := uint32(payload[4]) | uint32(payload[5])<<8 | uint32(payload[6])<<16 | uint32(payload[7])<<24
// 			unixf := float64(unix)
// 			msf := float64(ms) / 1000
// 			//valstring = fmt.Sprintf("%d.%d", unix, ms)
			

// 			valstring = fmt.Sprintf("%g", float64(unixf+msf))
// 			last_clock = valstring
// 		} else if field.Type == "byte" {
// 			sub := payload[field.Place[0]:field.Place[1]]
// 			if dbg {
// 				fmt.Printf("byte detected ")
// 				fmt.Printf(" sub  %x \n", sub)
// 			}
// 			valstring = string(sub)

// 		} else if field.Type == "int8_t" {
// 			if dbg {
// 				fmt.Printf("int 8 detected ")
// 			}
// 			sub := payload[field.Place[0]]
// 			if dbg {
// 				fmt.Printf(" sub  %x \n", sub)
// 			}
// 			data2 := int8(sub)
// 			tmpf := field.Factor * float64(data2)
// 			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)

// 		} else if field.Type == "uint8_t" {
// 			if dbg {
// 				fmt.Printf("uint 8 detected ")
// 			}
// 			sub := payload[field.Place[0]]
// 			if dbg {
// 				fmt.Printf(" sub  %x \n", sub)
// 			}
// 			data2 := sub
// 			tmpf := field.Factor * float64(data2)
// 			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)

// 		} else if field.Type == "int16_t" {
// 			if dbg {
// 				fmt.Printf("int 16 detected ")
// 			}
// 			sub := payload[field.Place[0]:field.Place[1]]
// 			if dbg {
// 				fmt.Printf(" sub  %x %x \n", sub[0], sub[1])
// 			}
// 			data2 := int16(sub[0]) | int16(sub[1])<<8

// 			tmpf := field.Factor * float64(data2)

// 			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)

// 		} else if field.Type == "uint16_t" {
// 			if dbg {
// 				fmt.Printf("uint 16 detected ")
// 			}
// 			sub := payload[field.Place[0]:field.Place[1]]
// 			if dbg {
// 				fmt.Printf(" sub  %x %x \n", sub[0], sub[1])
// 			}
// 			data2 := uint16(sub[0]) | uint16(sub[1])<<8

// 			tmpf := field.Factor * float64(data2)

// 			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)

// 		} else if field.Type == "int32_t" {
// 			if dbg {
// 				fmt.Printf("int 32 detected ")
// 			}
// 			sub := payload[field.Place[0]:field.Place[1]]
// 			if dbg {
// 				fmt.Printf(" sub  %x %x %x %x\n", sub[0], sub[1], sub[2], sub[3])
// 			}
// 			data2 := int32(sub[3]) | int32(sub[2])<<8 | int32(sub[1])<<16 | int32(sub[0])<<24

// 			tmpf := field.Factor * float64(data2)

// 			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)

// 		} else if field.Type == "uint32_t" {
// 			if dbg {
// 				fmt.Printf("uint 32 detected ")
// 			}
// 			sub := payload[field.Place[0]:field.Place[1]]
// 			if dbg {
// 				fmt.Printf(" sub  %x %x %x %x\n", sub[0], sub[1], sub[2], sub[3])
// 			}
// 			data2 := uint32(sub[3]) | uint32(sub[2])<<8 | uint32(sub[1])<<16 | uint32(sub[0])<<24

// 			tmpf := field.Factor * float64(data2)

// 			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
// 		} else if field.Type == "float" {
// 			if dbg {
// 				fmt.Printf("float 32 detected ")
// 			}
// 			sub := payload[field.Place[0]:field.Place[1]]
// 			if dbg {
// 				fmt.Printf(" sub  %x %x %x %x\n", sub[0], sub[1], sub[2], sub[3])
// 			}
// 			data3 := uint32(sub[0]) | uint32(sub[1])<<8 | uint32(sub[2])<<16 | uint32(sub[3])<<24
// 			data2 := math.Float32frombits(data3)
// 			tmpf := field.Factor * float64(data2)

// 			valstring = strconv.FormatFloat(tmpf, 'f', 5, 32)
// 		}

// 		retstr = retstr + "\"" + field.Key + "\" : " + valstring + ", "
// 	}
// 	if topic != "clock" {
// 		retstr = retstr + "\"unixtime\" : " + last_clock
// 	}
// 	if strings.HasSuffix(retstr, ", ") {
// 		retstr = strings.TrimRight(retstr, ", ")
// 	}
// 	retstr = retstr + "}"
// 	res := mqtt_response{}
// 	res.Topic = topic
// 	res.Payload = retstr
// 	return res
// }

// func convert2CAN(topic, payload string) CAN.CANFrame {
func convert2CAN(topic, payload string) can.Frame {
	conv, canid := getPayloadconv(&config, topic, "mqtt2can")

	fmt.Println(conv)
	//var data map[string]json.RawMessage
	var data map[string]interface{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		fmt.Println(err)
	}

	var buffer [8]uint8

	for _, field := range conv.Fields {
		if dbg {
			//fmt.Println("Key to find ", field.Key)
		}
		for key, value := range data {
			if dbg {
				//fmt.Println("Key:", key, " val ", value)
			}
			if key == field.Key {
				if dbg {
					fmt.Printf(" found Pair, going to parse %s %s %s\n", key, field.Key, field.Type)
				}
				if field.Type == "byte" {
					b, err := value.(string)
					if !err {
						fmt.Println("error converting value to byte")
					}
					for i := 0; i < len(b); i++ {
						buffer[i+field.Place[0]] = b[i] // write b into data starting at byte 2
					}
				} else if field.Type == "unixtime" {
					// do nothing
				} else {
					f, ok := value.(float64)

					if ok {
						//fmt.Println(f)
					} else {
						if dbg {
							fmt.Println("value is not a float64 setting to 0")
						}
						f = 0
					}
					//fmt.Printf("Type of %f  is %T   %f is %T\n", value, value, field.Factor, field.Factor)
					if field.Type == "uint8_t" {
						f64 := f * field.Factor
						u8 := uint8(f64)
						buffer[field.Place[0]] = u8
						if dbg {
							fmt.Printf("f %f to %f to int %d \n", f, f64, u8)
						}

					} else if field.Type == "int8_t" {
						f64 := f * field.Factor
						i8 := int8(f64)
						buffer[field.Place[0]] = byte(i8)
						if dbg {
							fmt.Printf("f %f to %f to int %d \n", f, f64, i8)
						}

					} else if field.Type == "uint16_t" {
						f64 := f * field.Factor
						u16 := uint16(f64)
						b := make([]byte, 2)
						binary.BigEndian.PutUint16(b, u16)
						for i := 0; i < len(b); i++ {
							buffer[i+field.Place[0]] = b[i] // write b into data starting at byte 2
						}
						if dbg {
							fmt.Printf("f %f to %f to int %d \n", f, f64, u16)
						}

					} else if field.Type == "int16_t" {
						f64 := f * field.Factor
						i16 := int16(f64)
						b := make([]byte, 2)
						binary.BigEndian.PutUint16(b, uint16(i16))
						for i := 0; i < len(b); i++ {
							buffer[i+field.Place[0]] = b[i] // write b into data starting at byte 2
						}
						if dbg {
							fmt.Printf("f %f to %f to int %d \n", f, f64, i16)
						}
					} else if field.Type == "uint32_t" {
						f64 := f * field.Factor
						u32 := uint32(f64)
						b := make([]byte, 4)
						binary.BigEndian.PutUint32(b, u32)
						for i := 0; i < len(b); i++ {
							buffer[i+field.Place[0]] = b[i] // write b into data starting at byte 2
						}
						if dbg {
							fmt.Printf("f %f to %f to int %d \n", f, f64, u32)
						}
					} else if field.Type == "int32_t" {
						f64 := f * field.Factor
						i32 := int32(f64)
						b := make([]byte, 4)
						binary.BigEndian.PutUint32(b, uint32(i32))
						for i := 0; i < len(b); i++ {
							buffer[i+field.Place[0]] = b[i] // write b into data starting at byte 2
						}
						if dbg {
							fmt.Printf("f %f to %f to int %d \n", f, f64, i32)
						}
					} else if field.Type == "float" {
						f64 := f * field.Factor
						f32 := float32(f64)
						b := make([]byte, 4)
						binary.BigEndian.PutUint32(b, math.Float32bits(f32))
						for i := 0; i < len(b); i++ {
							buffer[i+field.Place[0]] = b[i] // write b into data starting at byte 2
						}
						if dbg {
							fmt.Printf("f %f to %f to flo %f \n", f, f64, f32)
						}
					} else {
						fmt.Println("error conv instruction")
					}
				}

			}

		}
	}
	return_frame := can.Frame{
		ID:     0xFF,
		Length: 8,
		Flags:  0,
		Res0:   0,
		Res1:   0,
		Data:   [8]uint8{},
	}
	canidnr, err := strconv.ParseUint(canid, 0, 32)
	if err != nil {
		fmt.Println(err)
	}
	return_frame.ID = uint32(canidnr)
	return_frame.Data = buffer
	return return_frame
}
