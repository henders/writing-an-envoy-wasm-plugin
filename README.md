![Test Status](https://github.com/henders/writing-an-envoy-wasm-plugin/actions/workflows/test.yml/badge.svg)

# Envoy WASM Plugin
For Creating JWTs for service-to-service API requests

### Deploying to local K8s cluster

Build the Docker Image:
```shell
$ docker buildx build . -t shender/wasmplugin:v1
$ docker push shender/wasmplugin:v1
```

Deploy to K8s:
```shell
$ kubectl apply -f k8s_deploy.yml
```
