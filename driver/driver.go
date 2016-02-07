package driver


/*
#cgo LDFLAGS: -lcomedi -lpthread -lm
#cgo CFLAGS: -std=c99
#include "io.h"
#include "channels.h"
#include "elev.h"
*/
import "C"

type elevator_dir int
type button_type int

const (
	DOWN elevator_dir = -1
	STOP elevator_dir = 0
	UP elevator_dir = 1
)

func ElevatorInit(){
	C.elev_init()
}

func SetElevatorDir(dir elevator_dir){
	C.elev_set_motor_direction(C.elev_motor_direction_t(dir))
}


func SetDisconnectLamp(value int){
	C.elev_set_stop_lamp(C.int(value))
}

func SetButtonLamp(button button_type, floor int, value int){
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor),C.int(value))
}

func SetDoorOpen(value int ){
	C.elev_set_door_open_lamp(C.int(value))
}

func SetFloorIndicator(floor int){
	C.elev_set_floor_indicator(C.int(floor))
}

func GetFloorSensor() int{
	return int(C.elev_get_floor_sensor_signal())
}

func GetButtonSignal(button button_type, floor int) int{
	return int(C.elev_get_button_signal(C.elev_button_type_t(button),C.int(floor)))
}



