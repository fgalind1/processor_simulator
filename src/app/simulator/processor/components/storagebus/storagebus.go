package storagebus

import (
	"app/simulator/processor/models/operation"
)

type StorageBus struct {
	LoadRegister  func(uint32) uint32
	StoreRegister func(*operation.Operation, uint32, uint32)

	LoadData  func(uint32) uint32
	StoreData func(*operation.Operation, uint32, uint32)

	IncrementProgramCounter func(*operation.Operation, int32)
	SetProgramCounter       func(*operation.Operation, uint32)

	SetBranchResult func(uint32, bool)
}
