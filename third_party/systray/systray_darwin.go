package systray

/*
#cgo darwin CFLAGS: -DDARWIN -x objective-c -fobjc-arc
#cgo darwin LDFLAGS: -framework Cocoa -framework WebKit

#include <stdlib.h>
#include "systray.h"
*/
import "C"

import (
	"unsafe"
)

// StatusSegment is one inline avatar + text chunk for the menu bar title.
type StatusSegment struct {
	Text  string
	Image []byte
}

// SetStatusSegments renders avatar images inline with text in the menu bar title.
func SetStatusSegments(segments []StatusSegment) {
	n := len(segments)
	if n == 0 {
		return
	}

	segSize := C.size_t(unsafe.Sizeof(C.status_segment_t{}))
	base := C.malloc(segSize * C.size_t(n))
	if base == nil {
		return
	}
	defer C.free(base)

	csegs := unsafe.Slice((*C.status_segment_t)(base), n)
	textPtrs := make([]unsafe.Pointer, n)
	imgPtrs := make([]unsafe.Pointer, n)
	defer func() {
		for _, p := range textPtrs {
			if p != nil {
				C.free(p)
			}
		}
		for _, p := range imgPtrs {
			if p != nil {
				C.free(p)
			}
		}
	}()

	for i, s := range segments {
		if s.Text != "" {
			p := C.CString(s.Text)
			textPtrs[i] = unsafe.Pointer(p)
			csegs[i].text = p
		}
		if len(s.Image) > 0 {
			p := C.CBytes(s.Image)
			imgPtrs[i] = p
			csegs[i].image_bytes = (*C.char)(p)
			csegs[i].image_len = C.int(len(s.Image))
		}
	}

	C.setStatusSegments((*C.status_segment_t)(base), C.int(n))
}

// SetTemplateIcon sets the systray icon as a template icon (on Mac), falling back
// to a regular icon on other platforms.
// templateIconBytes and regularIconBytes should be the content of .ico for windows and
// .ico/.jpg/.png for other platforms.
func SetTemplateIcon(templateIconBytes []byte, regularIconBytes []byte) {
	cstr := (*C.char)(unsafe.Pointer(&templateIconBytes[0]))
	C.setIcon(cstr, (C.int)(len(templateIconBytes)), true)
}

// SetIcon sets the icon of a menu item. Only works on macOS and Windows.
// iconBytes should be the content of .ico/.jpg/.png
func (item *MenuItem) SetIcon(iconBytes []byte) {
	cstr := (*C.char)(unsafe.Pointer(&iconBytes[0]))
	C.setMenuItemIcon(cstr, (C.int)(len(iconBytes)), C.int(item.id), false)
}

// SetTemplateIcon sets the icon of a menu item as a template icon (on macOS). On Windows, it
// falls back to the regular icon bytes and on Linux it does nothing.
// templateIconBytes and regularIconBytes should be the content of .ico for windows and
// .ico/.jpg/.png for other platforms.
func (item *MenuItem) SetTemplateIcon(templateIconBytes []byte, regularIconBytes []byte) {
	cstr := (*C.char)(unsafe.Pointer(&templateIconBytes[0]))
	C.setMenuItemIcon(cstr, (C.int)(len(templateIconBytes)), C.int(item.id), true)
}
