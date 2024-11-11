BINARY_NAME=godis
MAIN_VER=1.2.9

DIST_WINDOWS=_dist/${BINARY_NAME}.exe
DIST_LINUX=_dist/${BINARY_NAME}
DIST_ARM64=_dist/${BINARY_NAME}-arm64
BUILD_DATE=`date`
LDFLAGS="-s -w -X 'main.builddate=${BUILD_DATE}' -X 'main.version=${MAIN_VER}'"

release: windows linux arm64
	@echo "copy files to server..."
	@scp -p ${DIST_WINDOWS} wlstl:/home/shares/archiving/v5release/luwakInstall/micro-services/bin
	@scp -p ${DIST_LINUX} wlstl:/home/shares/archiving/v5release/luwak_linux/bin
	@scp -p ${DIST_ARM64} wlstl:/home/shares/archiving/v5release/luwak_arm64/bin
	@echo "\nall done."

linux: modtidy
	@echo "building linux amd64 version..."
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o ${DIST_LINUX} -ldflags=${LDFLAGS} main.go
	@upx ${DIST_LINUX}
	@echo "done.\n"

windows: modtidy
	@echo "building windows amd64 version..."
	@GOARCH=amd64 GOOS=windows CGO_ENABLED=0 go build -o ${DIST_WINDOWS} -ldflags=${LDFLAGS} main.go
	@echo "done.\n"

arm64: modtidy
	@echo "building linux arm64/aarch64 version..."
	@GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -o ${DIST_ARM64} -ldflags=${LDFLAGS} main.go
	@echo "done.\n"

modtidy:
	@go mod tidy

push:
	@git gc
	@git fsck
	@git push