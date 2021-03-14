### Chat App

This is a sample chat application written in Go, using [websocket](https://github.com/gorilla/websocket) and [webrtc](https://webrtc.org/).

![](https://i.imgur.com/c0uwcbe.png)

### Development

```shell
# STEP 1: Clone this repository
git clone https://github.com/dongnguyenltqb/chatapp.git
cd chatapp
# STEP 2: Install dependency
go mod download
# STEP 3: Edit enviroment variable
vi .env
# STEP 4: Build and run application
make dev
```

### Deployment

```shell
# STEP 1: Build docker image
make build
# STEP 2: Delivery image to Dockerhub
make delivery
# STEP 3: Deloy application to k8s cluster
make deploy
```

