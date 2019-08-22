
all: frontend backend

frontend:
	npm run build

backend:
	cd pkg; GO111MODULE=on GOOS=linux GOARCH=amd64 go build -mod=vendor -o ../dist/grafana-scalyr-datasource-plugin_linux_amd64 .
	cd pkg; GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -mod=vendor -o ../dist/grafana-scalyr-datasource-plugin_darwin_amd64 .

first_pull:
	npm install
	cd pkg && dep ensure
