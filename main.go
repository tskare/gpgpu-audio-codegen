package main

import (
	"embed"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

//go:embed metal_gpuclass.h
//go:embed metal_gpuclass.moverride
//go:embed ext_sample_config.yaml
//go:embed cuda_gpuclass.cppoverride
//go:embed cuda_gpuclass.h
var embedfs embed.FS

const VERSION = "0.1.0"

type Buffer struct {
	Name       string `yaml:"name"`
	Size       string `yaml:"size"`
	Location   string `yaml:"location"`
	Parameters struct {
		Codegen struct {
			Init bool `yaml:"init"`
		} `yaml:"codegen"`
	} `yaml:"parameters"`
}

type Config struct {
	Parameters struct {
		Classname  string `yaml:"classname"`
		Samplerate int    `yaml:"samplerate"`
	} `yaml:"parameters"`
	Buffers []Buffer `yaml:"buffers"`
}

var flagInputFile *string
var flagOutputDir *string
var flagPlatforms *string
var flagRegenerateSourceNames *string

func parseFlagsToPointers() {
	flagInputFile = flag.String("input-config", "", "Input config file")
	flagOutputDir = flag.String("output-directory", "/tmp/gpuaudiogen", "Output directory for source files")
	flagPlatforms = flag.String("platforms", "", "Comma-separated list of platforms: {CUDA, Metal}. Default or empty is all.")
	flagVersion := flag.Bool("version", false, "Print version and exit")
	flagRegenerateSourceNames = flag.String("regenerate-source-mappings", "", "Regenerate files, original=current mappings, comma-separated, no spaces.")
	flag.Parse()

	if *flagVersion {
		fmt.Println("gpuaudiogen version " + VERSION)
		os.Exit(0)
	}
}

func getRequestedPlatforms(platforms string) []string {
	if platforms == "" {
		return []string{"CUDA", "Metal"}
	}
	return strings.Split(platforms, ",")
}

// CLEANUP: Use a built-in search function, but this is a go pattern.
func platformSearch(platforms []string, search string) bool {
	for _, platform := range platforms {
		if platform == search {
			return true
		}
	}
	return false
}
func includeCUDA(platforms []string) bool {
	return platformSearch(platforms, "CUDA")
}
func includeMetal(platforms []string) bool {
	return platformSearch(platforms, "Metal")
}
func replaceTokenInString(input string, token string, replacement string) string {
	return strings.ReplaceAll(input, token, replacement)
}
func getFileRegenerateMappingUfProvided(input string) string {
	if flagRegenerateSourceNames == nil || *flagRegenerateSourceNames == "" {
		return ""
	}
	mappings := strings.Split(*flagRegenerateSourceNames, ",")
	for _, mapping := range mappings {
		if strings.Contains(mapping, input) {
			return strings.Split(mapping, "=")[1]
		}
	}
	return ""
}

var CUDA_FILES = []string{"cuda_gpuclass.h", "cuda_gpuclass.cppoverride"}
var METAL_FILES = []string{"metal_gpuclass.h", "metal_gpuclass.moverride"}

func loadOnDiskFileToString(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error closing file: %+v", err)
		}
	}(file)

	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func loadEmbeddedFileToString(filePath string) (string, error) {
	filebytes, err := embedfs.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(filebytes), nil
}

func replaceRegexInString(input string, regex string, replacement string) string {
	// CLEANUP: cache these or move this to parent.
	re := regexp.MustCompile("(?s)" + regex)
	return re.ReplaceAllString(input, replacement)
}

// CLEANUP: Use a template engine.
func buildCudaBufferInit(config Config) string {
	bufferInit := ""
	sizeBytes := 4
	bufferInit += "//<GPUAGEN_CUDA_BUFFER_INIT>\n"

	for _, buffer := range config.Buffers {
		if buffer.Location == "device" {
			bufferInit += fmt.Sprintf("err = cudaMallocManaged(&%s, %s*%d);\n", buffer.Name, buffer.Size, sizeBytes)
			bufferInit += fmt.Sprintf("if(err != cudaSuccess) { printf(\"Error allocating %s: %s\\n\", cudaGetErrorString(err)); exit(EXIT_FAILURE); }\n\n", buffer.Name, buffer.Size)
		}
		if buffer.Location == "shared" {
			bufferInit += fmt.Sprintf("err =  cudaMallocManaged(&%s, %s*%d);\n", buffer.Name, buffer.Size, sizeBytes)
			bufferInit += fmt.Sprintf("if(err != cudaSuccess) { printf(\"Error allocating %s: %s\\n\", cudaGetErrorString(err)); exit(EXIT_FAILURE); }\n\n", buffer.Name, buffer.Size)
		}
		// FIXME
		if buffer.Location == "hybrid" {
			slog.Warn("Hybrid buffer not yet checked for correctness.")
			bufferInit += fmt.Sprintf("err = cudaMallocManaged(&%s, %s*%d);\n", buffer.Name, buffer.Size, sizeBytes)
			bufferInit += fmt.Sprintf("if(err != cudaSuccess) { printf(\"Error allocating %s: %s\\n\", cudaGetErrorString(err)); exit(EXIT_FAILURE); }\n\n", buffer.Name, buffer.Size)
		}
		if buffer.Location == "host" {
			bufferInit += fmt.Sprintf(
				"err = cudaMallocHost((void**)%s, %s * %d);", buffer.Name, buffer.Size, sizeBytes)

		}
	}
	bufferInit += "//<END_GPUAGEN_CUDA_BUFFER_INIT>\n"
	return bufferInit
}

func buildCUDABufferDefinitions(config Config) string {
	bufferDefs := ""
	bufferDefs += "//<GPUAGEN_CUDA_BUFFER_HEADER>\n"
	for _, buffer := range config.Buffers {
		if buffer.Location == "device" {
			bufferDefs += fmt.Sprintf("float *%s;\n", buffer.Name)
		}
		if buffer.Location == "shared" {
			bufferDefs += fmt.Sprintf("float *%s;\n", buffer.Name)
		}
		if buffer.Location == "hybrid" {
			bufferDefs += fmt.Sprintf("float *%s;\n", buffer.Name)
		}
		if buffer.Location == "host" {
			bufferDefs += fmt.Sprintf("float *%s;\n", buffer.Name)
		}
	}
	bufferDefs += "//<END_GPUAGEN_CUDA_BUFFER_HEADER>\n"
	return bufferDefs
}
func buildMetalBufferDefinitions(config Config) string {
	bufferDefs := ""
	bufferDefs += "//<GPUAGEN_METAL_BUFFER_HEADER>\n"
	for _, buffer := range config.Buffers {
		if buffer.Location == "device" {
			bufferDefs += fmt.Sprintf("id<MTLBuffer> %s;\n", buffer.Name)
		}
		if buffer.Location == "shared" {
			bufferDefs += fmt.Sprintf("id<MTLBuffer> %s;\n", buffer.Name)
		}
		if buffer.Location == "hybrid" {
			bufferDefs += fmt.Sprintf("id<MTLBuffer> %s;\n", buffer.Name)
		}
		if buffer.Location == "host" {
			bufferDefs += fmt.Sprintf("float* %s;\n", buffer.Name)
		}
	}
	bufferDefs += "//<END_GPUAGEN_METAL_BUFFER_HEADER>\n"
	return bufferDefs
}
func buildMetalBufferInit(config Config) string {
	bufferInit := ""
	sizeBytes := 4
	bufferInit += "//<GPUAGEN_METAL_BUFFER_INIT>\n"

	for _, buffer := range config.Buffers {
		// We default to shared memory architecture.
		if buffer.Location == "device" {
			bufferInit += fmt.Sprintf("%s = [device newBufferWithLength:%s*%d options:MTLResourceStorageModeShared];\n", buffer.Name, buffer.Size, sizeBytes)
		}
		if buffer.Location == "shared" {
			bufferInit += fmt.Sprintf("%s = [device newBufferWithLength:%s*%d options:MTLResourceStorageModeShared];\n", buffer.Name, buffer.Size, sizeBytes)
		}
		if buffer.Location == "hybrid" {
			bufferInit += fmt.Sprintf("%s = [device newBufferWithLength:%s*%d options:MTLResourceStorageModeShared];\n", buffer.Name, buffer.Size, sizeBytes)
		}
		if buffer.Location == "host" {
			bufferInit += fmt.Sprintf("%s = malloc(%s*%d);\n", buffer.Name, buffer.Size, sizeBytes)
		}
	}
	bufferInit += "//<END_GPUAGEN_METAL_BUFFER_INIT>\n"
	return bufferInit
}
func processOneFile(config Config, filePath string) {
	// Open file
	fmt.Printf("Processing %s\n", filePath)
	redirect := getFileRegenerateMappingUfProvided(filePath)
	var fstr string
	var err error
	if redirect != "" {
		fstr, err = loadOnDiskFileToString(redirect)
		if err != nil {
			log.Fatalf("Error loading file %s: %+v", filePath, err)
		}
	} else {
		fstr, err = loadEmbeddedFileToString(filePath)
		if err != nil {
			log.Fatalf("Error loading embedded file %s: %+v", filePath, err)
		}
	}

	newContent := fstr
	// Regress replace code back to its token.
	newContent = replaceRegexInString(newContent, ""+
		"\\/\\/<GPUAGEN_CUDA_BUFFER_HEADER>.*\\/\\/<END_GPUAGEN_CUDA_BUFFER_HEADER>", "gpuagen.tok.BUFFER_HEADER")
	newContent = replaceRegexInString(newContent, ""+
		"\\/\\/<GPUAGEN_CUDA_BUFFER_INIT>.*\\/\\/<END_GPUAGEN_CUDA_BUFFER_INIT>", "gpuagen.tok.BUFFER_INIT")

	// CLEANUP: Document regenerating class names being sensitive to newlines. It works, but it is not super readable.
	// See the embedded *.cuh file(s) for reference.
	newContent = replaceRegexInString(newContent, ""+
		"\\/\\/<GPUAGEN_CUDA_NAME>.*\\/\\/<END_GPUAGEN_CUDA_NAME>", "gpuagen.tok.NAME")

	// Replace various tokens (shared)
	// Note missing underscores on this one, for a couple corner cases with niche preprocessors and certain expression forms.
	// Likely we can ignore those -- CLEANUP.
	newContent = replaceTokenInString(newContent, "gpuagen.tok.SAMPLERATE", fmt.Sprintf("%d", config.Parameters.Samplerate))
	newContent = replaceTokenInString(newContent, "__gpuagen.tok.NAME", config.Parameters.Classname)

	// Replace various tokens (CUDA)
	newContent = replaceTokenInString(newContent, "__gpuagen.tok.BUFFER_HEADER", buildCUDABufferDefinitions(config))
	newContent = replaceTokenInString(newContent, "__gpuagen.tok.BUFFER_INIT", buildCudaBufferInit(config))

	// Replace various tokens (Metal)
	newContent = replaceTokenInString(newContent, "__gpuagen.tok.metal.BUFFER_HEADER", buildMetalBufferDefinitions(config))
	newContent = replaceTokenInString(newContent, "__gpuagen.tok.metal.BUFFER_INIT", buildMetalBufferInit(config))

	// Large block of virtual sample memory
	//	id<MTLBuffer> _mBufSampleMemory;

	// Write file to disk
	if filePath == "cuda_gpuclass.h" {
		filePath = config.Parameters.Classname + ".h"
	}
	if filePath == "cuda_gpuclass.cppoverride" {
		filePath = config.Parameters.Classname + ".cpp"
	}
	if filePath == "metal_gpuclass.h" {
		filePath = config.Parameters.Classname + "_metal.h"
	}
	if filePath == "metal_gpuclass.moverride" {
		filePath = config.Parameters.Classname + ".m"
	}
	outPathStr := fmt.Sprintf("%s/%s", *flagOutputDir, filePath)
	log.Printf("Writing file to disk: %s\n", outPathStr)
	os.WriteFile(outPathStr, []byte(newContent), 0644)
}

func processFiles(config Config, platforms []string) {
	if includeCUDA(platforms) {
		for _, file := range CUDA_FILES {
			processOneFile(config, file)
		}
	}
	if includeMetal(platforms) {
		for _, file := range METAL_FILES {
			processOneFile(config, file)
		}
	}
}
func main() {
	parseFlagsToPointers()

	var databyte []byte
	var err error

	if *flagInputFile != "" {
		databyte, err = os.ReadFile(*flagInputFile)
	} else {
		databyte, err = embedfs.ReadFile("ext_sample_config.yaml")
	}

	// Parse config, default from sample config embedded in binary.
	var config Config
	err = yaml.Unmarshal([]byte(databyte), &config)
	if err != nil {
		log.Fatalf("YAML parse error: %+v", err)
	}
	fmt.Printf("Parsed config: %+v\n", config)

	processFiles(config, getRequestedPlatforms(*flagPlatforms))
}
