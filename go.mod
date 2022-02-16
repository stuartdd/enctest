module stuartdd.com/enctest

go 1.16

replace stuartdd.com/lib => ./lib

replace stuartdd.com/gui => ./gui

replace stuartdd.com/theme2 => ./theme2

replace stuartdd.com/pref => ./pref

// replace fyne.io/fyne/v2 => github.com/andydotxyz/fyne/v2 07e8477c8

require (
	fyne.io/fyne/v2 v2.1.2-rc2.0.20220205054620-919d8dd6749e
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/stuartdd2/JsonParser4go/parser v0.0.0-20220216184436-59195ed65cbd
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	stuartdd.com/gui v0.0.0-00010101000000-000000000000
	stuartdd.com/lib v0.0.0-00010101000000-000000000000
	stuartdd.com/pref v0.0.0
	stuartdd.com/theme2 v0.0.0-00010101000000-000000000000
)
