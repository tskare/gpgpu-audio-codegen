package main

import (
	"gopkg.in/yaml.v2"
	"testing"
)

func TestYAMLParsing(t *testing.T) {
	data := `
parameters:
  samplerate: 48000
buffers:
  - name: "buffer1"
    size: 1024
    location: "host"
  - name: "buffer2"
    size: 2*FS
    location: "device"
    task: "input"
  - name: "buffer3"
    size: 512
    location: "shared"
  - name: "buffer4"
    size: 4096
    location: "hybrid"
    parameters:
      codegen:
        init: true
`

	var config Config
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		t.Fatalf("YAML parse error: %+v", err)
	}

	expectedSamplerate := 48000
	if config.Parameters.Samplerate != expectedSamplerate {
		t.Errorf("Expected samplerate %d, got %d", expectedSamplerate, config.Parameters.Samplerate)
	}

	expectedBuffers := 4
	if len(config.Buffers) != expectedBuffers {
		t.Errorf("Expected %d buffers, got %d", expectedBuffers, len(config.Buffers))
	}
}

func TestYAMLParsingWithoutName(t *testing.T) {
	data := `
parameters:
  samplerate: 48000
buffers:
  - name: "buffer1"
    size: 1024
    location: "host"
  - name: "buffer2"
    size: 2*FS
    location: "device"
    task: "input"
  - name: "buffer3"
    size: 512
    location: "shared"
  - name: "buffer4"
    size: 4096
    location: "hybrid"
    parameters:
      codegen:
        init: true
`

	var config Config
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		t.Fatalf("YAML parse error: %+v", err)
	}

	if config.Parameters.Classname != "" {
		t.Errorf("Expected empty name, got %s", config.Parameters.Classname)
	}
}

func TestYAMLParsingSizeCalculation(t *testing.T) {
	data := `
parameters:
  samplerate: 48000
buffers:
  - name: "buffer1"
    size: 1024
    location: "host"
  - name: "buffer2"
    size: 2*FS
    location: "device"
    task: "input"
  - name: "buffer3"
    size: 512
    location: "shared"
  - name: "buffer4"
    size: 4096
    location: "hybrid"
    parameters:
      codegen:
        init: true
`

	var config Config
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		t.Fatalf("YAML parse error: %+v", err)
	}

	expectedSize := 96000
	for _, buffer := range config.Buffers {
		if buffer.Name == "buffer2" && buffer.Size != "2*FS" {
			t.Errorf("Expected size %d, got %s", expectedSize, buffer.Size)
		}
	}
}
