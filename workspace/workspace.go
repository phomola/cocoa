package workspace

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#if __has_feature(objc_arc)
	#error ARC isn't allowed
#endif
inline bool cocoa_open_file(char* cfile) {
	__auto_type file = [[NSString alloc] initWithCString: cfile encoding: NSUTF8StringEncoding];
	__auto_type r = [[NSWorkspace sharedWorkspace] openURL: [NSURL fileURLWithPath: file]];
	[file release];
	return r;
}
*/
import "C"
import "unsafe"

func Open(file string) bool {
	cstring := C.CString(file)
	defer C.free(unsafe.Pointer(cstring))
	return bool(C.cocoa_open_file(cstring))
}
