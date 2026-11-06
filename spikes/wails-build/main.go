package main

/*
#cgo pkg-config: gtk+-3.0 webkit2gtk-4.1
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>
*/
import "C"
import (
	"fmt"
)

func main() {
	major := C.webkit_get_major_version()
	minor := C.webkit_get_minor_version()
	micro := C.webkit_get_micro_version()
	fmt.Printf("WebKitGTK CGO Linkage Success: %d.%d.%d\n", major, minor, micro)
}
