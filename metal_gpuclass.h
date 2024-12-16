#ifndef __gpuagen.tok.NAME_h
#define __gpuagen.tok.NAME_h

#import <Foundation/Foundation.h>
#import <Metal/Metal.h>
// We're using a base class from the GPUAudio benchmark framework.
// A future revision will remove this dependency.
#import "GPUABenchmark.h"

NS_ASSUME_NONNULL_BEGIN

@interface
__gpuagen.tok.NAME
: GPUABenchmark
- (instancetype) initWithDevice: (id<MTLDevice>) device;
- (void) setup;
- (void) runBenchmark: (NSArray*) latencies;
- (NSString*) getKernelName;

/*
// Not currently used in the header, but placed for reference:
__gpuagen.tok.BUFFER_HEADER
*/

@end


NS_ASSUME_NONNULL_END

#endif /*
__gpuagen.tok.NAME_h
*/