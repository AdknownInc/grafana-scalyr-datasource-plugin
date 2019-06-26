
all: frontend backend

frontend:
	npm run build

backend:
	GOOS=linux GOARCH=amd64 go build -o ./dist/grafana-scalyr-datasource-plugin_linux_amd64 ./pkg
	GOOS=darwin GOARCH=amd64 go build -o ./dist/grafana-scalyr-datasource-plugin_darwin_amd64 ./pkg

first_pull:
	npm install
	cd pkg && dep ensure
