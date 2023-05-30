# https://github.com/istio-ecosystem/wasm-extensions/blob/master/doc/how-to-build-oci-images.md
FROM scratch
ADD main.wasm ./plugin.wasm
