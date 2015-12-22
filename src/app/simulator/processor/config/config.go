package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"app/simulator/processor/consts"
)

type Config struct {
	*config
}

type config struct {
	CyclePeriodMs uint32 `json:"cycle_period_ms"`

	RegistersMemorySize    uint32 `json:"registers_memory_size"`
	InstructionsMemorySize uint32 `json:"instructions_memory_size"`
	DataMemorySize         uint32 `json:"data_memory_size"`

	Pipelined           bool          `json:"pipelined"`
	BranchPredictorType PredictorType `json:"branch_predictor_type"`

	InstructionsFetchedPerCycle    uint32 `json:"instructions_fetched_per_cycle"`
	InstructionsQueue              uint32 `json:"instructions_queue"`
	InstructionsDecodedQueue       uint32 `json:"instructions_decoded_queue"`
	InstructionsDispatchedPerCycle uint32 `json:"instructions_dispatched_per_cycle"`
	InstructionsWrittenPerCycle    uint32 `json:"instructions_written_per_cycle"`
	ReservationStationEntries      uint32 `json:"reservation_station_entries"`
	ReorderBufferEntries           uint32 `json:"reorder_buffer_entries"`
	RegisterAliasTableEntries      uint32 `json:"register_alias_table_entries"`

	DecoderUnits   uint32 `json:"decoder_units"`
	BranchUnits    uint32 `json:"branch_units"`
	LoadStoreUnits uint32 `json:"load_store_units"`
	AluUnits       uint32 `json:"alu_units"`
	FpuUnits       uint32 `json:"fpu_units"`
}

type PredictorType string

const (
	// No predictor
	StallPredictor PredictorType = "stall"

	// Static predictors
	AlwaysTakenPredictor   PredictorType = "always_taken"
	NeverTakenPredictor    PredictorType = "never_taken"
	BackwardTakenPredictor PredictorType = "backward_taken"
	ForwardTakenPredictor  PredictorType = "forward_taken"

	// Dynamic predictors
	OneBitPredictor PredictorType = "one_bit"
	TwoBitPredictor PredictorType = "two_bit"
)

func Load(filename string) (*Config, error) {

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var c *Config
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil, err
	}

	return c, nil
}

func (this *Config) CyclePeriodMs() uint32 {
	return this.config.CyclePeriodMs
}

func (this *Config) CyclePeriod() time.Duration {
	return time.Duration(this.config.CyclePeriodMs) * time.Millisecond
}

func (this *Config) RegistersMemorySize() uint32 {
	return this.config.RegistersMemorySize
}

func (this *Config) TotalRegisters() uint32 {
	return this.RegistersMemorySize() / consts.BYTES_PER_WORD
}

func (this *Config) InstructionsMemorySize() uint32 {
	return this.config.InstructionsMemorySize
}

func (this *Config) DataMemorySize() uint32 {
	return this.config.DataMemorySize
}

func (this *Config) Pipelined() bool {
	return this.config.Pipelined
}

func (this *Config) BranchPredictorType() PredictorType {
	return this.config.BranchPredictorType
}

func (this *Config) InstructionsFetchedPerCycle() uint32 {
	return this.config.InstructionsFetchedPerCycle
}

func (this *Config) InstructionsQueue() uint32 {
	return this.config.InstructionsQueue
}

func (this *Config) InstructionsDecodedQueue() uint32 {
	return this.config.InstructionsDecodedQueue
}

func (this *Config) InstructionsDispatchedPerCycle() uint32 {
	return this.config.InstructionsDispatchedPerCycle
}

func (this *Config) InstructionsWrittenPerCycle() uint32 {
	return this.config.InstructionsWrittenPerCycle
}

func (this *Config) ReservationStationEntries() uint32 {
	return this.config.ReservationStationEntries
}

func (this *Config) ReorderBufferEntries() uint32 {
	return this.config.ReorderBufferEntries
}

func (this *Config) RegisterAliasTableEntries() uint32 {
	return this.config.RegisterAliasTableEntries
}

func (this *Config) DecoderUnits() uint32 {
	return this.config.DecoderUnits
}

func (this *Config) AluUnits() uint32 {
	return this.config.AluUnits
}

func (this *Config) FpuUnits() uint32 {
	return this.config.FpuUnits
}

func (this *Config) LoadStoreUnits() uint32 {
	return this.config.LoadStoreUnits
}

func (this *Config) BranchUnits() uint32 {
	return this.config.BranchUnits
}

func (this *Config) ToString() string {
	str := "\n Processor Config:\n\n"
	str += fmt.Sprintf(" => Cycle Period: %d ms\n", this.CyclePeriodMs())
	str += fmt.Sprintf(" => Bytes per word: %d\n", consts.BYTES_PER_WORD)
	str += fmt.Sprintf(" => Registers: %d\n", this.TotalRegisters())
	str += fmt.Sprintf(" => Instr Memory: %d Bytes\n", this.InstructionsMemorySize())
	str += fmt.Sprintf(" => Data Memory: %d Bytes\n", this.DataMemorySize())
	str += fmt.Sprintf(" => Pipelined: %v\n", this.Pipelined())
	str += fmt.Sprintf(" => Branch Predictor Type: %v\n", this.BranchPredictorType())
	str += fmt.Sprintf(" => Instructions Fetched per Cycle: %d\n", this.InstructionsFetchedPerCycle())
	str += fmt.Sprintf(" => Instructions Queue (IQ): %d\n", this.InstructionsQueue())
	str += fmt.Sprintf(" => Instructions Decoded Queue (IDQ): %d\n", this.InstructionsDecodedQueue())
	str += fmt.Sprintf(" => Instructions Dispatched per Cycle: %d\n", this.InstructionsFetchedPerCycle())
	str += fmt.Sprintf(" => Instructions Written per Cycle: %d\n", this.InstructionsWrittenPerCycle())
	str += fmt.Sprintf(" => Reservation Station Entries: %d\n", this.ReservationStationEntries())
	str += fmt.Sprintf(" => Re-order Buffer Entries: %d\n", this.ReorderBufferEntries())
	str += fmt.Sprintf(" => Register Alias Table Entries: %d\n", this.RegisterAliasTableEntries())
	str += fmt.Sprintf(" => Decoder Units: %d\n", this.DecoderUnits())
	str += fmt.Sprintf(" => Alu Units: %d\n", this.AluUnits())
	str += fmt.Sprintf(" => FPU Units: %d\n", this.FpuUnits())
	str += fmt.Sprintf(" => Load/Store Units: %d\n", this.LoadStoreUnits())
	str += fmt.Sprintf(" => Branch Units: %d\n", this.BranchUnits())
	return str
}
