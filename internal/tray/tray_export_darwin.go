//go:build darwin

package tray

import "C"

//export goTrayStart
func goTrayStart() {
	if current != nil && current.start != nil {
		current.start()
	}
}

//export goTrayShow
func goTrayShow() {
	if current != nil && current.OnShow != nil {
		current.OnShow()
	}
}
