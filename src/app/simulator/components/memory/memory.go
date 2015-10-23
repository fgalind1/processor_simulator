package memory

type Memory struct {
	memory *memory
}

type memory struct {
	size uint32
	data []byte
}

func New(size uint32) *Memory {
	return &Memory{
		&memory{
			size: size,
			data: make([]byte, size),
		},
	}
}

func (this *Memory) Size() uint32 {
	return this.memory.size
}

func (this *Memory) Data() []byte {
	return this.memory.data
}

func (this *Memory) Load(address uint32, lenght uint32) []byte {
	return this.memory.data[address : address+lenght]
}

func (this *Memory) LoadUint32(address uint32) uint32 {
	return uint32(this.memory.data[address+3])<<24 +
		uint32(this.memory.data[address+2])<<16 +
		uint32(this.memory.data[address+1])<<8 +
		uint32(this.memory.data[address+0])<<0
}

func (this *Memory) Store(address uint32, values ...byte) {
	for i, value := range values {
		this.memory.data[address+uint32(i)] = value
	}
}

func (this *Memory) StoreUint32(address uint32, value uint32) {
	this.memory.data[address+3] = byte((value & 0xFF000000) >> 24)
	this.memory.data[address+2] = byte((value & 0x00FF0000) >> 16)
	this.memory.data[address+1] = byte((value & 0x0000FF00) >> 8)
	this.memory.data[address+0] = byte((value & 0x000000FF) >> 0)
}
