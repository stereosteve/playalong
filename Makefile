dev::
	wgo -file=.go -file=.templ -xfile=_templ.go templ generate :: go run server.go

client::
	wgo -file=.ts bun build.ts

ship::
	bun build.ts
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/jamfu-amd64
	ssh jam -t 'systemctl stop jamfu'
	scp build/jamfu-amd64 jam:jamfu
	ssh jam -t 'systemctl start jamfu'
