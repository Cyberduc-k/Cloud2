# Sudokus

## Setup
This app can be run in two ways:
1. Docker compose: run docker-compose, every microservice will be available on a different port. login: 8081, highscores: 8082, start: 8083, stop: 8084.
2. Minikube/kubernetes.

## Minikube setup
First you need to install minikube: [https://minikube.sigs.k8s.io/docs/start/](https://minikube.sigs.k8s.io/docs/start/).
Next you need to run a series of comands, these are differnt on linux and windows.

### Linux
The following command you only need to run once:
```sh
sudo groupadd docker
sudo usermod -aG docker $USER
```

Every time you want to run minikube you have to run the following:
```sh
newgrp docker
minikube docker-env | source
minikube start
```


### Windows
Every time you want to run minikube:
```sh
minikube docker-env | Invoke-Expression
minikube start
```

## Deploy
To deploy the app in minikube run the following:
```sh
docker-compose build
kubectl apply -f kubernetes.yaml
```

Optionally you can open the dashboard with:
```sh
minikube dashboard
```