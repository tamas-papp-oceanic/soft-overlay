package main

/*
#cgo pkg-config: gtk+-3.0 gtk-layer-shell-0
#include <gtk/gtk.h>
#include <gtk-layer-shell.h>
static void init_layer(GtkWindow *w) {
	gtk_layer_init_for_window(w);
	gtk_layer_set_layer(w, GTK_LAYER_SHELL_LAYER_OVERLAY);   // highest layer
	gtk_layer_auto_exclusive_zone_enable(w);                 // don’t shrink Chromium
	gtk_layer_set_anchor(w, GTK_LAYER_SHELL_EDGE_TOP,    TRUE);
	gtk_layer_set_anchor(w, GTK_LAYER_SHELL_EDGE_BOTTOM, TRUE);
	gtk_layer_set_anchor(w, GTK_LAYER_SHELL_EDGE_LEFT,   TRUE);
	gtk_layer_set_anchor(w, GTK_LAYER_SHELL_EDGE_RIGHT,  TRUE);
}
*/
import "C"
import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"unsafe"

	"github.com/gotk3/gotk3/gtk"
)

const backlightDir = "/sys/class/backlight"

func main() {
	gtk.Init(nil)

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetDecorated(false)
	win.Fullscreen()
	C.init_layer((*C.GtkWindow)(unsafe.Pointer(win.Native())))

	overlay, _ := gtk.OverlayNew()
	win.Add(overlay)

	// dimming surface
	bg, _ := gtk.DrawingAreaNew()
	bg.SetHExpand(true)
	bg.SetVExpand(true)
	css := `
	#dim-layer {
			background-color: rgba(0, 0, 0, 0.5);
	}
	`
	cssProvider, _ := gtk.CssProviderNew()
	cssProvider.LoadFromData(css)
	screen := win.GetScreen()
	gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	// vertical box in the middle
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 12)
	box.SetName("dim-layer")
	box.SetHAlign(gtk.ALIGN_CENTER)
	box.SetVAlign(gtk.ALIGN_CENTER)
	overlay.AddOverlay(box)

	icon, _ := gtk.ImageNewFromIconName("system-lock-screen", gtk.ICON_SIZE_DIALOG)
	box.PackStart(icon, false, false, 0)

	scale, _ := gtk.ScaleNewWithRange(gtk.ORIENTATION_HORIZONTAL, 1, 100, 1)
	scale.SetValue(100)
	box.PackStart(scale, false, false, 0)

	btn, _ := gtk.ButtonNewWithLabel("Apply brightness")
	box.PackStart(btn, false, false, 0)

	btn.Connect("clicked", func() {
		val := int(scale.GetValue())
		if err := setBrightness(val); err != nil {
			fmt.Println("brightness:", err)
		}
		// hide overlay after applying
		win.Destroy()
	})

	win.Connect("destroy", gtk.MainQuit)
	win.ShowAll()
	gtk.Main()
}

// func rgba(r, g, b, a float64) gdk.RGBA { return gdk.NewRGBA(r, g, b, a) }

// setBrightness writes percentage to the first backlight device it finds.
func setBrightness(p int) error {
	devices, _ := filepath.Glob(filepath.Join(backlightDir, "*"))
	if len(devices) == 0 {
		return fmt.Errorf("no backlight devices")
	}
	dev := devices[0]
	maxBytes, _ := ioutil.ReadFile(filepath.Join(dev, "max_brightness"))
	max, _ := strconv.Atoi(string(maxBytes[:len(maxBytes)-1]))
	value := max * p / 100
	return ioutil.WriteFile(filepath.Join(dev, "brightness"), []byte(fmt.Sprint(value)), 0664)
}
