package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"rclone-sync-toggle/internal/config"
	"rclone-sync-toggle/internal/i18n"
)

type managerProfile struct {
	profile *config.ConnectorProfile
	active  bool
}

type managerResult struct {
	action  string
	active  []string
	deleted []string
	err     error
}

type gtkTreeIter struct {
	stamp     int32
	padding   int32
	userData  uintptr
	userData2 uintptr
	userData3 uintptr
}

type gtkValue struct {
	gType uintptr
	data  [2]uintptr
}

type gtkAPI struct {
	gtkInitCheck                  func(unsafe.Pointer, unsafe.Pointer) int32
	gtkMain                       func()
	gtkMainQuit                   func()
	gtkWindowNew                  func(int32) uintptr
	gtkWindowSetTitle             func(uintptr, string)
	gtkWindowSetDefaultSize       func(uintptr, int32, int32)
	gtkContainerSetBorderWidth    func(uintptr, uint32)
	gtkContainerAdd               func(uintptr, uintptr)
	gtkBoxNew                     func(int32, int32) uintptr
	gtkBoxPackStart               func(uintptr, uintptr, int32, int32, uint32)
	gtkListStoreNewv              func(int32, *uintptr) uintptr
	gtkListStoreAppend            func(uintptr, *gtkTreeIter)
	gtkListStoreSetValue          func(uintptr, *gtkTreeIter, int32, *gtkValue)
	gtkListStoreRemove            func(uintptr, *gtkTreeIter) int32
	gtkTreeModelGetIterFromString func(uintptr, *gtkTreeIter, string) int32
	gtkTreeViewNewWithModel       func(uintptr) uintptr
	gtkTreeViewSetHeadersVisible  func(uintptr, int32)
	gtkTreeViewGetPathAtPos       func(uintptr, int32, int32, *uintptr, *uintptr, *int32, *int32) int32
	gtkTreeViewAppendColumn       func(uintptr, uintptr) int32
	gtkTreeViewColumnNew          func() uintptr
	gtkTreeViewColumnSetTitle     func(uintptr, string)
	gtkTreeViewColumnPackStart    func(uintptr, uintptr, int32)
	gtkTreeViewColumnAddAttribute func(uintptr, uintptr, string, int32)
	gtkTreeViewColumnSetExpand    func(uintptr, int32)
	gtkCellRendererToggleNew      func() uintptr
	gtkCellRendererTextNew        func() uintptr
	gtkCellRendererPixbufNew      func() uintptr
	gtkScrolledWindowNew          func(unsafe.Pointer, unsafe.Pointer) uintptr
	gtkScrolledWindowSetPolicy    func(uintptr, int32, int32)
	gtkButtonBoxNew               func(int32) uintptr
	gtkButtonBoxSetLayout         func(uintptr, int32)
	gtkButtonNewWithLabel         func(string) uintptr
	gtkWidgetSetSizeRequest       func(uintptr, int32, int32)
	gtkWidgetShowAll              func(uintptr)
	gtkWidgetDestroy              func(uintptr)
	gtkTimeoutAdd                 func(uint32, uintptr, uintptr) uint32
	gtkTreePathGetIndices         func(uintptr) *int32
	gtkTreePathFree               func(uintptr)
	gdkEventGetCoords             func(uintptr, *float64, *float64) int32
	gdkEventGetButton             func(uintptr, *uint32) int32
	gtkCellRendererGetType        func() uintptr
	gTypeFromName                 func(string) uintptr
	gTypeClassRef                 func(uintptr) uintptr
	gValueInit                    func(*gtkValue, uintptr) *gtkValue
	gValueSetBoolean              func(*gtkValue, int32)
	gValueSetString               func(*gtkValue, string)
	gValueSetEnum                 func(*gtkValue, int32)
	gValueUnset                   func(*gtkValue)
	gObjectSetProperty            func(uintptr, string, *gtkValue)
	gSignalConnectData            func(uintptr, string, uintptr, uintptr, uintptr, uint32) uintptr
}

var (
	gtkOnce      sync.Once
	gtk          gtkAPI
	gtkLoadError error
)

func showConfigManager(profiles []*config.ConnectorProfile) managerResult {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	gtkOnce.Do(loadGTK)
	if gtkLoadError != nil {
		return managerResult{err: gtkLoadError}
	}
	if gtk.gtkInitCheck(nil, nil) == 0 {
		return managerResult{err: fmt.Errorf("could not initialize GTK; check that a graphical session is available")}
	}

	rows := make([]managerProfile, 0, len(profiles))
	for _, profile := range profiles {
		rows = append(rows, managerProfile{profile: profile, active: profile.Enabled})
	}

	valueTypes := []uintptr{
		gtk.gTypeFromName("gboolean"),
		gtk.gTypeFromName("gchararray"),
		gtk.gTypeFromName("gchararray"),
		gtk.gTypeFromName("gchararray"),
		gtk.gTypeFromName("gchararray"),
	}
	for _, valueType := range valueTypes {
		if valueType == 0 {
			return managerResult{err: fmt.Errorf("GTK does not provide a required list column type")}
		}
	}
	model := gtk.gtkListStoreNewv(int32(len(valueTypes)), &valueTypes[0])

	for _, row := range rows {
		remoteName, cloudPath := splitRemote(row.profile.Remote)
		iter := gtkTreeIter{}
		gtk.gtkListStoreAppend(model, &iter)
		setGTKValue(model, &iter, 0, valueTypes[0], row.active)
		setGTKValue(model, &iter, 1, valueTypes[1], remoteName)
		setGTKValue(model, &iter, 2, valueTypes[2], displayPath(row.profile.LocalDir))
		setGTKValue(model, &iter, 3, valueTypes[3], cloudPath)
		setGTKValue(model, &iter, 4, valueTypes[4], row.profile.Mode)
	}

	window := gtk.gtkWindowNew(0)
	gtk.gtkWindowSetTitle(window, i18n.T("ui.manager_title"))
	gtk.gtkWindowSetDefaultSize(window, 1020, 440)
	gtk.gtkContainerSetBorderWidth(window, 14)

	layout := gtk.gtkBoxNew(1, 12)
	gtk.gtkContainerAdd(window, layout)

	tree := gtk.gtkTreeViewNewWithModel(model)
	gtk.gtkTreeViewSetHeadersVisible(tree, 1)

	activeRenderer := gtk.gtkCellRendererToggleNew()
	appendGTKColumn(tree, i18n.T("ui.active"), activeRenderer, "active", 0, false)

	for _, column := range []struct {
		title string
		index int32
		wide  bool
	}{
		{title: i18n.T("ui.connection"), index: 1},
		{title: i18n.T("ui.local_folder"), index: 2, wide: true},
		{title: i18n.T("ui.cloud_folder"), index: 3, wide: true},
		{title: i18n.T("ui.mode"), index: 4},
	} {
		renderer := gtk.gtkCellRendererTextNew()
		if column.wide {
			setGTKRendererProperty(renderer, "ellipsize", gtk.gTypeFromName("PangoEllipsizeMode"), 3)
		}
		appendGTKColumn(tree, column.title, renderer, "text", column.index, column.wide)
	}

	trashRenderer := gtk.gtkCellRendererPixbufNew()
	gtk.gTypeClassRef(gtk.gtkCellRendererGetType())
	setGTKRendererProperty(trashRenderer, "icon-name", valueTypes[1], "user-trash-symbolic")
	trashColumn := appendGTKColumn(tree, i18n.T("ui.delete"), trashRenderer, "icon-name", -1, false)
	scroll := gtk.gtkScrolledWindowNew(nil, nil)
	gtk.gtkScrolledWindowSetPolicy(scroll, 1, 1)
	gtk.gtkContainerAdd(scroll, tree)
	gtk.gtkBoxPackStart(layout, scroll, 1, 1, 0)

	buttons := gtk.gtkButtonBoxNew(0)
	gtk.gtkButtonBoxSetLayout(buttons, 3)
	addButton := gtk.gtkButtonNewWithLabel(i18n.T("ui.add_synced_folder"))
	saveButton := gtk.gtkButtonNewWithLabel(i18n.T("ui.save_apply"))
	gtk.gtkWidgetSetSizeRequest(addButton, 160, -1)
	gtk.gtkWidgetSetSizeRequest(saveButton, 120, -1)
	gtk.gtkContainerAdd(buttons, addButton)
	gtk.gtkContainerAdd(buttons, saveButton)
	gtk.gtkBoxPackStart(layout, buttons, 0, 0, 0)

	result := managerResult{action: "cancel"}
	windowDestroyed := false
	mainLoopRunning := true
	addGTKSignal(activeRenderer, "toggled", func(_ uintptr, path *byte, _ uintptr) {
		index := gtkPathIndex(path)
		if index < 0 || index >= len(rows) {
			return
		}
		rows[index].active = !rows[index].active
		iter, ok := gtkIterForPath(model, path)
		if ok {
			setGTKValue(model, &iter, 0, valueTypes[0], rows[index].active)
		}
	})
	addGTKSignal(tree, "button-press-event", func(_ uintptr, event uintptr, _ uintptr) int32 {
		var button uint32
		if gtk.gdkEventGetButton(event, &button) == 0 || button != 1 {
			return 0
		}
		var x, y float64
		if gtk.gdkEventGetCoords(event, &x, &y) == 0 {
			return 0
		}
		var path, column uintptr
		var cellX, cellY int32
		if gtk.gtkTreeViewGetPathAtPos(tree, int32(x), int32(y), &path, &column, &cellX, &cellY) == 0 {
			return 0
		}
		defer gtk.gtkTreePathFree(path)
		if column != trashColumn {
			return 0
		}
		indices := gtk.gtkTreePathGetIndices(path)
		if indices == nil {
			return 0
		}
		index := int(*indices)
		if index < 0 || index >= len(rows) {
			return 0
		}
		result.deleted = append(result.deleted, rows[index].profile.Name)
		iter := gtkTreeIter{}
		if gtk.gtkTreeModelGetIterFromString(model, &iter, strconv.Itoa(index)) != 0 {
			gtk.gtkListStoreRemove(model, &iter)
		}
		rows = append(rows[:index], rows[index+1:]...)
		return 1
	})
	addGTKSignal(addButton, "clicked", func(_ uintptr, _ uintptr) {
		result.action = "add"
		gtk.gtkWidgetDestroy(window)
	})
	addGTKSignal(saveButton, "clicked", func(_ uintptr, _ uintptr) {
		result.action = "save"
		gtk.gtkWidgetDestroy(window)
	})
	addGTKSignal(window, "destroy", func(_ uintptr, _ uintptr) {
		windowDestroyed = true
		if mainLoopRunning {
			gtk.gtkMainQuit()
		}
	})

	gtk.gtkWidgetShowAll(window)
	gtk.gtkMain()
	mainLoopRunning = false
	if !windowDestroyed {
		gtk.gtkWidgetDestroy(window)
	}
	for _, row := range rows {
		if row.active {
			result.active = append(result.active, row.profile.Name)
		}
	}
	return result
}

func loadGTK() {
	library, err := purego.Dlopen("libgtk-3.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		gtkLoadError = fmt.Errorf("could not load GTK 3: %w", err)
		return
	}
	register := func(target any, name string) {
		if gtkLoadError != nil {
			return
		}
		symbol, err := purego.Dlsym(library, name)
		if err != nil {
			gtkLoadError = fmt.Errorf("could not load GTK function %s: %w", name, err)
			return
		}
		purego.RegisterFunc(target, symbol)
	}

	register(&gtk.gtkInitCheck, "gtk_init_check")
	register(&gtk.gtkMain, "gtk_main")
	register(&gtk.gtkMainQuit, "gtk_main_quit")
	register(&gtk.gtkWindowNew, "gtk_window_new")
	register(&gtk.gtkWindowSetTitle, "gtk_window_set_title")
	register(&gtk.gtkWindowSetDefaultSize, "gtk_window_set_default_size")
	register(&gtk.gtkContainerSetBorderWidth, "gtk_container_set_border_width")
	register(&gtk.gtkContainerAdd, "gtk_container_add")
	register(&gtk.gtkBoxNew, "gtk_box_new")
	register(&gtk.gtkBoxPackStart, "gtk_box_pack_start")
	register(&gtk.gtkListStoreNewv, "gtk_list_store_newv")
	register(&gtk.gtkListStoreAppend, "gtk_list_store_append")
	register(&gtk.gtkListStoreSetValue, "gtk_list_store_set_value")
	register(&gtk.gtkListStoreRemove, "gtk_list_store_remove")
	register(&gtk.gtkTreeModelGetIterFromString, "gtk_tree_model_get_iter_from_string")
	register(&gtk.gtkTreeViewNewWithModel, "gtk_tree_view_new_with_model")
	register(&gtk.gtkTreeViewSetHeadersVisible, "gtk_tree_view_set_headers_visible")
	register(&gtk.gtkTreeViewGetPathAtPos, "gtk_tree_view_get_path_at_pos")
	register(&gtk.gtkTreeViewAppendColumn, "gtk_tree_view_append_column")
	register(&gtk.gtkTreeViewColumnNew, "gtk_tree_view_column_new")
	register(&gtk.gtkTreeViewColumnSetTitle, "gtk_tree_view_column_set_title")
	register(&gtk.gtkTreeViewColumnPackStart, "gtk_tree_view_column_pack_start")
	register(&gtk.gtkTreeViewColumnAddAttribute, "gtk_tree_view_column_add_attribute")
	register(&gtk.gtkTreeViewColumnSetExpand, "gtk_tree_view_column_set_expand")
	register(&gtk.gtkCellRendererToggleNew, "gtk_cell_renderer_toggle_new")
	register(&gtk.gtkCellRendererTextNew, "gtk_cell_renderer_text_new")
	register(&gtk.gtkCellRendererPixbufNew, "gtk_cell_renderer_pixbuf_new")
	register(&gtk.gtkScrolledWindowNew, "gtk_scrolled_window_new")
	register(&gtk.gtkScrolledWindowSetPolicy, "gtk_scrolled_window_set_policy")
	register(&gtk.gtkButtonBoxNew, "gtk_button_box_new")
	register(&gtk.gtkButtonBoxSetLayout, "gtk_button_box_set_layout")
	register(&gtk.gtkButtonNewWithLabel, "gtk_button_new_with_label")
	register(&gtk.gtkWidgetSetSizeRequest, "gtk_widget_set_size_request")
	register(&gtk.gtkWidgetShowAll, "gtk_widget_show_all")
	register(&gtk.gtkWidgetDestroy, "gtk_widget_destroy")
	register(&gtk.gtkTimeoutAdd, "g_timeout_add")
	register(&gtk.gtkTreePathGetIndices, "gtk_tree_path_get_indices")
	register(&gtk.gtkTreePathFree, "gtk_tree_path_free")
	register(&gtk.gdkEventGetCoords, "gdk_event_get_coords")
	register(&gtk.gdkEventGetButton, "gdk_event_get_button")
	register(&gtk.gtkCellRendererGetType, "gtk_cell_renderer_get_type")
	register(&gtk.gTypeFromName, "g_type_from_name")
	register(&gtk.gTypeClassRef, "g_type_class_ref")
	register(&gtk.gValueInit, "g_value_init")
	register(&gtk.gValueSetBoolean, "g_value_set_boolean")
	register(&gtk.gValueSetString, "g_value_set_string")
	register(&gtk.gValueSetEnum, "g_value_set_enum")
	register(&gtk.gValueUnset, "g_value_unset")
	register(&gtk.gObjectSetProperty, "g_object_set_property")
	register(&gtk.gSignalConnectData, "g_signal_connect_data")
}

func appendGTKColumn(tree uintptr, title string, renderer uintptr, attribute string, index int32, expand bool) uintptr {
	column := gtk.gtkTreeViewColumnNew()
	gtk.gtkTreeViewColumnSetTitle(column, title)
	gtk.gtkTreeViewColumnPackStart(column, renderer, 1)
	if index >= 0 {
		gtk.gtkTreeViewColumnAddAttribute(column, renderer, attribute, index)
	}
	expandValue := int32(0)
	if expand {
		expandValue = 1
	}
	gtk.gtkTreeViewColumnSetExpand(column, expandValue)
	gtk.gtkTreeViewAppendColumn(tree, column)
	return column
}

func setGTKValue(model uintptr, iter *gtkTreeIter, column int32, valueType uintptr, value any) {
	var valueBuffer gtkValue
	gtk.gValueInit(&valueBuffer, valueType)
	switch value := value.(type) {
	case bool:
		boolValue := int32(0)
		if value {
			boolValue = 1
		}
		gtk.gValueSetBoolean(&valueBuffer, boolValue)
	case string:
		gtk.gValueSetString(&valueBuffer, value)
	}
	gtk.gtkListStoreSetValue(model, iter, column, &valueBuffer)
	gtk.gValueUnset(&valueBuffer)
}

func setGTKRendererProperty(renderer uintptr, name string, valueType uintptr, value any) {
	if valueType == 0 {
		return
	}
	var valueBuffer gtkValue
	gtk.gValueInit(&valueBuffer, valueType)
	switch value := value.(type) {
	case string:
		gtk.gValueSetString(&valueBuffer, value)
	case int:
		gtk.gValueSetEnum(&valueBuffer, int32(value))
	}
	gtk.gObjectSetProperty(renderer, name, &valueBuffer)
	gtk.gValueUnset(&valueBuffer)
}

func addGTKSignal(instance uintptr, signal string, callback any) {
	pointer := purego.NewCallback(callback)
	gtk.gSignalConnectData(instance, signal, pointer, 0, 0, 0)
}

func gtkPathIndex(path *byte) int {
	index, err := strconv.Atoi(strings.TrimSpace(gtkCString(path)))
	if err != nil {
		return -1
	}
	return index
}

func gtkIterForPath(model uintptr, path *byte) (gtkTreeIter, bool) {
	iter := gtkTreeIter{}
	ok := gtk.gtkTreeModelGetIterFromString(model, &iter, gtkCString(path)) != 0
	return iter, ok
}

func gtkCString(pointer *byte) string {
	if pointer == nil {
		return ""
	}
	var value []byte
	for offset := uintptr(0); ; offset++ {
		character := *(*byte)(unsafe.Add(unsafe.Pointer(pointer), offset))
		if character == 0 {
			return string(value)
		}
		value = append(value, character)
	}
}

func splitRemote(remote string) (string, string) {
	name, path, found := strings.Cut(remote, ":")
	if !found {
		return remote, "/"
	}
	if path == "" {
		path = "/"
	}
	return name, path
}

func displayPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return filepath.Clean(path)
}
