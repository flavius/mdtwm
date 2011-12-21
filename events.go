package main

import (
	"code.google.com/p/x-go-binding/xgb"
)

func handleEvent(event xgb.Event) {
	switch e := event.(type) {
	// Request events
	case xgb.MapRequestEvent:
		mapRequest(e)
	case xgb.ConfigureRequestEvent:
		configureRequest(e)

	// Notify events
	case xgb.EnterNotifyEvent:
		enterNotify(e)
	case xgb.MapNotifyEvent:
		mapNotify(e)
	case xgb.UnmapNotifyEvent:
		unmapNotify(e)
	case xgb.DestroyNotifyEvent:
		destroyNotify(e)
	case xgb.ConfigureNotifyEvent:
		configureNotify(e)

	// Keyboard and mouse events
	case xgb.KeyPressEvent:
		keyPress(e)
	case xgb.ButtonPressEvent:
		buttonPress(e)
	case xgb.ButtonReleaseEvent:
		buttonRelease(e)
	case xgb.MotionNotifyEvent:
		motionNotify(e)
	default:
		d.Printf("Unhandled event: %T: %+v", e, e)
	}
}

func mapRequest(e xgb.MapRequestEvent) {
	d.Printf("%T: %+v", e, e)
	manage(Window(e.Window) , currentPanel(), false)
}

// For windows with override-redirect flag
func mapNotify(e xgb.MapNotifyEvent) {
	d.Printf("%T: %+v", e, e)
	manage(Window(e.Window) , currentPanel(), false)
}

func enterNotify(e xgb.EnterNotifyEvent) {
	d.Printf("%T: %+v", e, e)
	if e.Mode != xgb.NotifyModeNormal {
		return
	}
	w := Window(e.Event)
	currentDesk.SetFocus(currentDesk.Window() == w)
	// Iterate over all boxes in current desk
	bi := currentDesk.Children().FrontIter(true)
	for b := bi.Next(); b != nil; b = bi.Next() {
		b.SetFocus(b.Window() == w)
	}
	statusLog()
}

func destroyNotify(e xgb.DestroyNotifyEvent) {
	d.Printf("%T: %+v", e, e)
	unmanage(Window(e.Window))
}

func unmapNotify(e xgb.UnmapNotifyEvent) {
	d.Printf("%T: %+v", e, e)
	// We mask UnmapNotify during reparenting, so these events have
	// e.Event == root. root isn't managed so unamange(root) do nothing.
	unmanage(Window(e.Event))
}

func configureNotify(e xgb.ConfigureNotifyEvent) {
	d.Printf("%T: %+v", e, e)
	b := root.Children().BoxByWindow(Window(e.Event), true)
	if b == nil {
		return
	}
	b.SyncGeometry(Geometry{
		e.X, e.Y,
		Int16(e.Width), Int16(e.Height),
		Int16(e.BorderWidth),
	})
}

func configureRequest(e xgb.ConfigureRequestEvent) {
	d.Printf("%T: %+v", e, e)
	w := Window(e.Window)
	b := root.Children().BoxByWindow(w, true)
	if b == nil || b.Float() {
		// We accept request from floating windows.
		// Unmanaged window will be configured by manage() function so
		// now we can simply execute its request.
		mask := (xgb.ConfigWindowX | xgb.ConfigWindowY |
			xgb.ConfigWindowWidth | xgb.ConfigWindowHeight |
			xgb.ConfigWindowBorderWidth | xgb.ConfigWindowSibling |
			xgb.ConfigWindowStackMode) & e.ValueMask
		v := make([]interface{}, 0, 7)
		if mask&xgb.ConfigWindowX != 0 {
			v = append(v, e.X)
		}
		if mask&xgb.ConfigWindowY != 0 {
			v = append(v, e.Y)
		}
		if mask&xgb.ConfigWindowWidth != 0 {
			v = append(v, e.Width)
		}
		if mask&xgb.ConfigWindowHeight != 0 {
			v = append(v, e.Height)
		}
		if mask&xgb.ConfigWindowBorderWidth != 0 {
			v = append(v, e.BorderWidth)
		}
		if mask&xgb.ConfigWindowSibling != 0 {
			v = append(v, e.Sibling)
		}
		if mask&xgb.ConfigWindowStackMode != 0 {
			v = append(v, e.StackMode)
		}
		w.Configure(mask, v...)
		return
	}
	// Force box configuration
	g := b.Geometry()
	cne := xgb.ConfigureNotifyEvent{
		Event:        e.Window,
		Window:       e.Window,
		AboveSibling: xgb.WindowNone,
		X:            g.X,
		Y:            g.Y,
		Width:        Pint16(g.W),
		Height:       Pint16(g.H),
		BorderWidth:  Pint16(g.B),
	}
	w.Send(false, xgb.EventMaskStructureNotify, cne)
}