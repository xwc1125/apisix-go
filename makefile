# 设置编译后文件的名称
BINARY=apisix-go
# stable,beta,alpha
VERSION_META=beta
# 编译号文件名称
BUILD_FILE_PATH=./
BUILD_VERSION_FILE=Version
BUILD_NUMBER_FILE=BuildNumber
# 抓取当前git最新的hash码
GIT_HASH=`git rev-parse HEAD`
GIT_DATE=`git show --pretty=format:"%ct" | head -1`
# 编译日期
BUILD=`date +%FT%T%z`
# 文件输出的文件夹
OutputVersion=`./${BINARY} out-version-meta`
OutputDir=./app_package/

# 设置go程序中，对应变量的值
LDFLAGS=-ldflags "-w -s  \
-X github.com/xwc1125/apisix-go/pkg/version.FilePath=$(BUILD_FILE_PATH)  \
-X github.com/xwc1125/apisix-go/pkg/version.Version=$$(cat $(BUILD_FILE_PATH)$(BUILD_VERSION_FILE))  \
-X github.com/xwc1125/apisix-go/pkg/version.VersionMeta=${VERSION_META}  \
-X github.com/xwc1125/apisix-go/pkg/version.BuildTime=${BUILD}  \
-X github.com/xwc1125/apisix-go/pkg/version.GitHash=${GIT_HASH}  \
-X github.com/xwc1125/apisix-go/pkg/version.GitDate=${GIT_DATE}  \
-X github.com/xwc1125/apisix-go/pkg/version.BuildNumber=$$(cat $(BUILD_FILE_PATH)$(BUILD_NUMBER_FILE))"

mkdir:
	$(shell if [ ! -e $(BUILD_FILE_PATH) ];then mkdir -p $(BUILD_FILE_PATH); fi)
	$(shell if [ ! -e $(OutputDir) ];then mkdir -p $(OutputDir); fi)

build: mkdir
	@go run ${LDFLAGS} ${BUILD_FILE_PATH} build
	@go build ${LDFLAGS} -o ${APP_NAME} ${BUILD_FILE_PATH}
	./${APP_NAME} version
	@rm $(BUILD_FILE_PATH)${BUILD_NUMBER_FILE}
	@echo "==================>"
rebuild: mkdir
	@go run ${LDFLAGS} ${BUILD_FILE_PATH} rebuild
	@go build ${LDFLAGS} -o ${APP_NAME} ${BUILD_FILE_PATH}
	./${APP_NAME} version
	@rm $(BUILD_FILE_PATH)${BUILD_NUMBER_FILE}
	@echo "==================>"

build-linux: mkdir build
	@echo "start build version:v${OutputVersion}"
	# linux
	@GOOS=linux CGO_ENABLED=0 go build ${LDFLAGS} -o ${BINARY}_linux_v${OutputVersion}
	@mv ${BINARY}_linux_v${OutputVersion} ${OutputDir}
	@echo "build linux success"
	# delete file
	@rm $(BUILD_FILE_PATH)${BUILD_NUMBER_FILE}
	@rm ${BINARY}
build-all: mkdir build
	@echo "start build version:v${OutputVersion}"
	# linux
	@GOOS=linux CGO_ENABLED=0 go build ${LDFLAGS} -o ${BINARY}_linux_v${OutputVersion}
	@mv ${BINARY}_linux_v${OutputVersion} ${OutputDir}
	@echo "build linux success"
	# mac
	@GOOS=darwin CGO_ENABLED=0 go build ${LDFLAGS} -o ${BINARY}_darwin_v${OutputVersion}
	@mv ${BINARY}_darwin_v${OutputVersion} ${OutputDir}
	@echo "build darwin success"
	# windows
	@GOOS=windows CGO_ENABLED=0 go build ${LDFLAGS} -o ${BINARY}_windows_v${OutputVersion}
	@mv ${BINARY}_windows_v${OutputVersion} ${OutputDir}
	@echo "build windows success"
	# delete file
	@rm $(BUILD_FILE_PATH)${BUILD_NUMBER_FILE}
	@rm ${BINARY}

rebuild-linux: mkdir rebuild
	@echo "start build version:v${OutputVersion}"
	# linux
	@GOOS=linux CGO_ENABLED=0 go build ${LDFLAGS} -o ${APP_NAME}_linux_v${OutputVersion} ${BUILD_FILE_PATH}
	@mv ${APP_NAME}_linux_v${OutputVersion} ${OutputDir}
	@echo "build linux success"
	# delete file
	@rm $(BUILD_FILE_PATH)${BUILD_NUMBER_FILE}
	@rm ${APP_NAME}
rebuild-all: mkdir rebuild
	@echo "start build version:v${OutputVersion}"
	# linux
	@GOOS=linux CGO_ENABLED=0 go build ${LDFLAGS} -o ${APP_NAME}_linux_v${OutputVersion} ${BUILD_FILE_PATH}
	@mv ${APP_NAME}_linux_v${OutputVersion} ${OutputDir}
	@echo "build linux success"
	# mac
	@GOOS=darwin CGO_ENABLED=0 go build ${LDFLAGS} -o ${APP_NAME}_darwin_v${OutputVersion} ${BUILD_FILE_PATH}
	@mv ${APP_NAME}_darwin_v${OutputVersion} ${OutputDir}
	@echo "build darwin success"
	# windows
	@GOOS=windows CGO_ENABLED=0 go build ${LDFLAGS} -o ${APP_NAME}_windows_v${OutputVersion} ${BUILD_FILE_PATH}
	@mv ${APP_NAME}_windows_v${OutputVersion} ${OutputDir}
	@echo "build windows success"
	# delete file
	@rm $(BUILD_FILE_PATH)${BUILD_NUMBER_FILE}
	@rm ${APP_NAME}


