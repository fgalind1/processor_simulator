package config

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type Config struct {
	*config
}

type config struct {
	CyclePeriodMs uint32 `json:"cycle_period_ms"`

	RegistersMemorySize    uint32 `json:"registers_memory_size"`
	InstructionsMemorySize uint32 `json:"instructions_memory_size"`
	DataMemorySize         uint32 `json:"data_memory_size"`

	InstructionsBufferSize uint32 `json:"instructions_buffer_size"`
	DecoderUnits           uint32 `json:"decoder_units"`
	BranchUnits            uint32 `json:"branch_units"`
	LoadStoreUnits         uint32 `json:"load_store_units"`
	AluUnits               uint32 `json:"alu_units"`
}

func New(cyclePeriodMs, registersMemorySize, instructionsMemorySize, dataMemorySize,
	instructionsBufferSize, decoderUnits, branchUnits, loadStoreUnits, aluUnits uint32) *Config {
	return &Config{
		&config{
			CyclePeriodMs:          cyclePeriodMs,
			RegistersMemorySize:    registersMemorySize,
			InstructionsMemorySize: instructionsMemorySize,
			DataMemorySize:         dataMemorySize,

			InstructionsBufferSize: instructionsBufferSize,
			DecoderUnits:           decoderUnits,
			BranchUnits:            branchUnits,
			LoadStoreUnits:         loadStoreUnits,
			AluUnits:               aluUnits,
		},
	}
}

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

func (this *Config) CyclePeriodMs() time.Duration {
	return time.Duration(this.config.CyclePeriodMs) * time.Millisecond
}

func (this *Config) RegistersMemorySize() uint32 {
	return this.config.RegistersMemorySize
}

func (this *Config) InstructionsMemorySize() uint32 {
	return this.config.InstructionsMemorySize
}

func (this *Config) DataMemorySize() uint32 {
	return this.config.DataMemorySize
}

func (this *Config) InstructionsBufferSize() uint32 {
	return this.config.InstructionsBufferSize
}

func (this *Config) DecoderUnits() uint32 {
	return this.config.DecoderUnits
}

func (this *Config) AluUnits() uint32 {
	return this.config.AluUnits
}

func (this *Config) LoadStoreUnits() uint32 {
	return this.config.LoadStoreUnits
}

func (this *Config) BranchUnits() uint32 {
	return this.config.BranchUnits
}
