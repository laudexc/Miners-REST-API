package logic

import "errors"

var (
	ErrNotStarted           = errors.New("enterprise is not started")
	ErrAlreadyStarted       = errors.New("enterprise already started")
	ErrAlreadyStopped       = errors.New("enterprise already stopped")
	ErrUnknownMinerClass    = errors.New("unknown miner class")
	ErrMinersWrongQuantity  = errors.New("the number of miners to hire cannot be less than 1")
	ErrUnknownEquipmentType = errors.New("unknown equipment type")
	ErrNotEnoughCoal        = errors.New("not enough coal")
	ErrEquipmentBought      = errors.New("equipment already bought")
)
