package storagebus

type StorageBus struct {
	LoadRegister  func(uint32) uint32
	StoreRegister func(uint32, uint32)

	LoadData  func(uint32) uint32
	StoreData func(uint32, uint32)

	SetProgramCounter       func(uint32)
	IncrementProgramCounter func(int32)
}
