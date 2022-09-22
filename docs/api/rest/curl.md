# Laniakea REST API - curl

In this guide, we will look at how to interact with the Laniakea REST API via curl

## Guide

Once you have an instance of `laniakea` running and unlocked, grab the location of the `tls.cert` file and `admin.macaroon` file. Let's first declare a variable in our bash terminal window
```
$ MACAROON_HEADER="Grpc-Metadata-macaroon: $(xxd -ps -u -c 1000 /home/user/.laniakea/admin.macaroon)"
```
Now we make the call
```
$ curl -X GET --cacert /home/user/.laniakea/tls.cert --header "$MACAROON_HEADER" https://localhost:8080/v1/test
{"msg":"This is a regular test"}
```
We can also make streaming RPC calls with curl since our reverse proxy implements WebSockets.
```
$ curl -X GET --cacert /home/user/.laniakea/tls.cert --header "$MACAROON_HEADER" https://localhost:8080/v2/plugin/subscribe/rng-plugin
{"result":{"source":"demo-datasource-plugin","type":"application/json","timestamp":"1663888625063","payload":"eyJkYXRhI..."}}
```