module stuartdd.com/enctest

go 1.16

replace stuartdd.com/lib => ./lib

replace stuartdd.com/gui => ./gui

replace stuartdd.com/theme2 => ./theme2

replace stuartdd.com/pref => ./pref

replace stuartdd.com/types => ./types

// replace fyne.io/fyne/v2 => github.com/andydotxyz/fyne/v2 07e8477c8

require (
	fyne.io/fyne/v2 v2.1.0
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-gl/gl v0.0.0-20210905235341-f7a045908259 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20210727001814-0db043d8d5be // indirect
	github.com/godbus/dbus/v5 v5.0.5 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/stuartdd/jsonParserGo/parser v0.0.0-20211019172056-78b407826a60
	golang.org/x/sys v0.0.0-20211020064051-0ec99a608a1b // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	stuartdd.com/gui v0.0.0-00010101000000-000000000000
	stuartdd.com/lib v0.0.0-00010101000000-000000000000
	stuartdd.com/pref v0.0.0
	stuartdd.com/theme2 v0.0.0-00010101000000-000000000000
	stuartdd.com/types v0.0.0
)
