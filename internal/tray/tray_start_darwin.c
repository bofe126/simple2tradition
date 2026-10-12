// Dispatches the systray start onto the main GCD queue: AppKit
// requires NSStatusItem (created inside goTrayStart via fyne systray)
// to be instantiated on the main thread.
#include <dispatch/dispatch.h>

extern void goTrayStart(void);

void dispatchTrayStart(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		goTrayStart();
	});
}
