# CRUD gRPC Template [Golang + GORM] Server&Client

### Sources
- https://grpc.io/blog/coreos/
- https://github.com/philips/grpc-gateway-example/
- https://dev.to/chilladx/how-we-use-grpc-to-build-a-clientserver-system-in-go-1mi
- https://gorm.io/

#### Tags
- add custom tags by https://github.com/favadi/protoc-go-inject-tag

## ENVS

- DBCONN    - connection string
- TBL       - database table

- GRPC      - grpc port
- REST      - rest port
- UI        - ui port
 
## RUN 

1. Clone project

```
git clone https://github.com/FilipAnteKovacic/grpc-gorm-template.git
```

2. Create or copy, modifiy proto file

- define messages
- define services

3. Make sure you have google proto files

```
https://github.com/googleapis/api-common-protos
```

4. Server name change

- generate.sh
```
SERVER_CN=localhost - change
```

- server.go
```
creds.OverrideServerName("localhost") - change
```

5. Run generate.sh

```
./generate.sh
```

6. Change itemData, reqDataToStuct, structDataToRes

7. Run app

```
DBCONN="user:pass@(localhost)/grpc?charset=utf8&parseTime=True&loc=Local" TBL="data" UI="7000" REST="7010" GRPC="7020" go run *.go
```

## Docker

1. Build image

```
docker build -t grpc:gorm .
```

2. Run container

```
docker run -d  -e "DBCONN=user:pass@(localhost)/grpc?charset=utf8&parseTime=True&loc=Local" -e "TBL=grpc" -e "UI=8060" -e "REST=8070" -e "GRPC=8080" -p 8060:8060 -p 8070:8070 -p 8080:8080 --name grpc-gorm-template grpc:gorm
```