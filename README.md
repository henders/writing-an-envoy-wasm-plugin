![Test Status](https://github.com/henders/writing-an-envoy-wasm-plugin/actions/workflows/test.yml/badge.svg)

# Envoy WASM Plugin
For retrieving JWTs for service-to-service API requests. This code matches Part 5 of the Writing an Istio WASM Plugin in Go for migrating 100s of services to new auth strategy Medium article.
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
