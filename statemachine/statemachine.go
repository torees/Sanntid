package statemachine

type State int

type elevStates struct{
	State idle = 0
	State down = -1
	State up = 1
	State doorOpen = 2
}

