// Injects an applicationShouldHandleReopen: handler into the Wails
// app delegate so clicking the Dock icon re-shows the hidden window.
// Wails v2 does not implement this delegate method itself, so we add
// it at runtime with class_addMethod.
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

extern void goTrayShow(void);

static BOOL handleReopen(id self, SEL _cmd, NSApplication *app, BOOL hasVisibleWindows) {
	goTrayShow();
	return NO;
}

// Swaps the Dock icon at runtime (direction-aware branding). NSData
// copies the buffer synchronously, so the Go slice may be freed as
// soon as this returns.
void setDockIcon(const void* bytes, int len) {
	NSData *data = [NSData dataWithBytes:bytes length:len];
	dispatch_async(dispatch_get_main_queue(), ^{
		NSImage *img = [[NSImage alloc] initWithData:data];
		if (img != nil) {
			[NSApp setApplicationIconImage:img];
		}
	});
}

void installReopenHandler(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		id delegate = [NSApp delegate];
		if (delegate == nil) {
			return;
		}
		Class cls = object_getClass(delegate);
		SEL sel = @selector(applicationShouldHandleReopen:hasVisibleWindows:);
		if (!class_respondsToSelector(cls, sel)) {
			class_addMethod(cls, sel, (IMP)handleReopen, "B@:@B");
		}
	});
}
