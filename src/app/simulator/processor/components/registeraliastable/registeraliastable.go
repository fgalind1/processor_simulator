package registeraliastable

const INVALID = -1

type RegisterAliasTable struct {
	*registerAliasTable
}

type registerAliasTable struct {
	index   uint32
	entries []RatEntry
}

type RatEntry struct {
	Free         bool
	ArchRegister uint32
	OperationId  int32
}

func New(index uint32, entries uint32) *RegisterAliasTable {
	rat := &RegisterAliasTable{
		&registerAliasTable{
			index:   index,
			entries: make([]RatEntry, entries),
		},
	}
	for entryIndex, _ := range rat.Entries() {
		rat.Entries()[entryIndex] = RatEntry{
			Free: true,
		}
	}
	return rat
}

func (this *RegisterAliasTable) Index() uint32 {
	return this.registerAliasTable.index
}

func (this *RegisterAliasTable) Entries() []RatEntry {
	return this.registerAliasTable.entries
}

func (this *RegisterAliasTable) AddMap(destination, operationId uint32) (bool, uint32) {
	found, index := this.getNextFreeEntry()
	if !found {
		return false, 0
	}

	this.Entries()[index] = RatEntry{
		Free:         false,
		ArchRegister: destination,
		OperationId:  int32(operationId),
	}
	return true, index
}

func (this *RegisterAliasTable) Release(operationId uint32) {
	for entryIndex, _ := range this.Entries() {
		if this.registerAliasTable.entries[entryIndex].OperationId == int32(operationId) {
			this.registerAliasTable.entries[entryIndex] = RatEntry{
				Free: true,
			}
		}
	}
}

func (this *RegisterAliasTable) GetPhysicalRegister(operationId uint32, archRegister uint32) (uint32, bool) {
	found := false
	physicalIndex := int32(INVALID)
	physicalReg := RatEntry{
		OperationId: INVALID,
	}

	for entryIndex, entry := range this.Entries() {
		if !entry.Free && entry.ArchRegister == archRegister && entry.OperationId > physicalReg.OperationId && uint32(entry.OperationId) <= operationId {
			physicalIndex = int32(entryIndex)
			physicalReg = entry
			found = true
		}
	}

	return uint32(physicalIndex), found
}

func (this *RegisterAliasTable) getNextFreeEntry() (bool, uint32) {
	for index, entry := range this.Entries() {
		if entry.Free {
			return true, uint32(index)
		}
	}
	return false, 0
}
