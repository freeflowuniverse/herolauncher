module github.com/freeflowuniverse/herolauncher/pkg/vfs/9p2000

go 1.23.0

toolchain go1.24.1

require (
	github.com/freeflowuniverse/herolauncher v0.0.0-20250315180128-b9a3b6627b56
	github.com/knusbaum/go9p v1.18.0
)

require (
	9fans.net/go v0.0.2 // indirect
	github.com/Plan9-Archive/libauth v0.0.0-20180917063427-d1ca9e94969d // indirect
	github.com/emersion/go-sasl v0.0.0-20220912192320-0145f2c60ead // indirect
	github.com/fhs/mux9p v0.3.1 // indirect
)

replace github.com/knusbaum/go9p => ../../../../../knusbaum/go9p
