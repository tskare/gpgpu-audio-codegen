# GPGPU Audio Code Generator

This utility assists in creating or updating boilerplate for kernels that may process audio on the GPU.

It may save some time if you expect to need to resize buffers frequently, or for a first run.
Otherwise you may prefer to copy some reference code manually and proceed from there.

It may be used with the GPGPU Audio Benchmark framework sibling repository: 

or for your own programs.
It dovetails with the C++ audio transfer macros that will be placed in the util/ folder. (TODO: copy those in).

You may use the provided release binary or build using the Go toolkit; the only known dependency is the YAML toolkit.

## Usage

```bash
gpuacodegen [FLAGS] --inputFilename myConfig.yaml

FLAGS:
    -h, --help         Prints help information
    --version          Prints version information
    --inputFilename    The input YAML file to read from
    --outputDirectory  The output directory to write to
    --platforms        Comma-separated: e.g. CUDA,Metal
Advanced:
    --regenerate-source-mappings Comma-separated list of files to feed back into the tool.
                                 These should be of the format template_file=your_file. 
   
```

## Open items

Build files (Make/CMake/VSProj/etc.) are not generated until files are processed in bulk. 

  - For Metal, consider starting with the Apple "Performing Calculations on a GPU" example. 
  - For CUDA, consider starting with the VectorAdd example in the CUDA Toolkit `samples` folder.

Regenerating code after edits, via `regenerate-source-mappings`, is supported but not unit-tested or validated beyond basic workflows.
