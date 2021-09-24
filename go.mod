module stuartdd.com/enctest

go 1.16

replace stuartdd.com/lib => ./lib

replace stuartdd.com/gui => ./gui

replace stuartdd.com/theme2 => ./theme2

replace stuartdd.com/pref => ./pref

// replace fyne.io/fyne/v2 => github.com/andydotxyz/fyne/v2 07e8477c8

require (
	fyne.io/fyne/v2 v2.1.0
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20210727001814-0db043d8d5be // indirect
	github.com/srwiley/oksvg v0.0.0-20210519022825-9fc0c575d5fe // indirect
	github.com/srwiley/rasterx v0.0.0-20210519020934-456a8d69b780 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	stuartdd.com/gui v0.0.0-00010101000000-000000000000
	stuartdd.com/lib v0.0.0-00010101000000-000000000000
	stuartdd.com/pref v0.0.0-00010101000000-000000000000
	stuartdd.com/theme2 v0.0.0-00010101000000-000000000000
)
