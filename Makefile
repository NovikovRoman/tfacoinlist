FILENAME=tfacoinlist
BUILD_DIR=dist
ENV_FILE=${BUILD_DIR}/.env

build:
	mkdir -p ${BUILD_DIR}
	redoc-cli bundle docs/openapi/index.yaml -o ${BUILD_DIR}/openapi.html
	go build -ldflags="-s -w" -o ${BUILD_DIR}/${FILENAME}
	if [ ! -f "${ENV_FILE}" ]; then\
		cp .env-sample ${ENV_FILE};\
	fi