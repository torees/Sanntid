package driver


/*
#cgo LDFLAGS: -lcomedi -lpthread -lm
#cgo CFLAGS: -std=c99
#include "io.h"
#include "channels.h"
#include "elev.h"
*/
import "C"

type Elevator_dir int
type Button_type int

const (
	DOWN Elevator_dir = -1
	STOP Elevator_dir = 0
	UP Elevator_dir = 1
)

func ElevatorInit(){
	C.elev_init()
}

func ElevatorStart(dir Elevator_dir){
	C.elev_set_motor_direction(C.elev_motor_direction_t(dir))
}


func NetworkDisconnect(value int){
	C.elev_set_stop_lamp(C.int(value))
}

func ButtonLamp(button Button_type, floor int, value int){
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor),C.int(value))
}

func DoorOpen(value int ){
	C.elev_set_door_open_lamp(C.int(value))
}

func FloorIndicator(floor int){
	C.elev_set_floor_indicator(C.int(floor))
}

func FloorSensor() int{
	return int(C.elev_get_floor_sensor_signal())
}

func ButtonPushed(button Button_type, floor int) int{
	return int(C.elev_get_button_signal(C.elev_button_type_t(button),C.int(floor)))
}



