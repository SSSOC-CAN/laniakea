# Laniakea gRPC API - Python

In this guide, we will look at how to interact with the Laniakea gRPC API via Python

## Dependencies

Create a new virtual environment and install the following packages

```
$ python -m venv venv
$ source venv/bin/activate
(venv) $ pip install --upgrade pip
(venv) $ pip install grpcio grpcio-tools
(venv) $ touch grpc_test.py
```
Create a blank python file, the name doesn't matter. We'll call ours `grpc_test.py`. 

## Generating stubs

Create a new folder called `protos` and copy the `lani.proto`, `plugin.proto` and `pluginapi.proto` files into this folder from `lanirpc` [folder](https://github.com/SSSOC-CAN/fmtd/blob/main/lanirpc)
```
(venv) $ mkdir protos & cd protos
(venv) $ curl -o lani.proto -s https://raw.githubusercontent.com/SSSOC-CAN/fmtd/main/lanirpc/lani.proto
(venv) $ curl -o plugin.proto -s https://raw.githubusercontent.com/SSSOC-CAN/fmtd/main/lanirpc/plugin.proto
(venv) $ curl -o pluginapi.proto -s https://raw.githubusercontent.com/SSSOC-CAN/fmtd/main/lanirpc/pluginapi.proto
```
Once there, we must generate the stubs from the proto file.
```
(venv) $ python -m grpc_tools.protoc --proto_path=googleapis:. --python_out=. --grpc_python_out=. plugin.proto
(venv) $ python -m grpc_tools.protoc --proto_path=googleapis:. --python_out=. --grpc_python_out=. lani.proto
(venv) $ python -m grpc_tools.protoc --proto_path=googleapis:. --python_out=. --grpc_python_out=. pluginapi.proto
```
Six new files should be generated `lani_pb2.py`, `lani_pb2_grpc.py`, `plugin_pb2.py`, `plugin_pb2_grpc.py`, `pluginapi_pb2.py` and `pluginapi_pb2_grpc.py`. Edit the `lani_pb2.py`, `lani_pb2_grpc.py`, `plugin_pb2_grpc.py`, `pluginapi_pb2.py` and `pluginapi_pb2_grpc.py` files and change the import statements to add the `protos` folder.
```python
import protos.lani_pb2 as lani__pb2
import protos.pluginapi_pb2 as pluginapi__pb2
import protos.plugin_pb2 as plugin__pb2
```

## Write the test script

Return to the main folder and edit the main file. In order to use the gRPC API, we need the Laniakea TLS certificate file and macaroon. We read these files in our script to establish the connection to Laniakea via gRPC.

```python
import protos.lani_pb2 as lani
import protos.lani_pb2_grpc as lanirpc
import protos.pluginapi_pb2 as plugin
import protos.pluginapi_pb2_grpc as pluginrpc
import grpc
import os
import codecs

# Due to updated ECDSA generated tls.cert we need to let gprc know that
# we need to use that cipher suite otherwise there will be a handhsake
# error when we communicate with the lnd rpc server.
os.environ["GRPC_SSL_CIPHER_SUITES"] = 'HIGH+ECDSA'

# Lnd cert is at ~/.laniakea/tls.cert on Linux and
# ~/Library/Application Support/Laniakea/tls.cert on Mac
# C:\Users\Your User\AppData\Local\Laniakea on Windows
cert = open(os.path.expanduser('~/.laniakea/tls.cert'), 'rb').read()
creds = grpc.ssl_channel_credentials(cert)
channel = grpc.secure_channel('localhost:7777', creds)
lani_stub = lanirpc.LaniStub(channel)
plugin_stub = pluginrpc.PluginAPIStub(channel)

# Laniakea admin macaroon is at ~/.laniakea/admin.macaroon on Linux and
# ~/Library/Application Support/Laniakea/admin.macaroon on Mac and
# C:\Users\Your User\AppData\Local\Laniakea\admin.macaroon on Windows
with open(os.path.expanduser('~/.laniakea/admin.macaroon'), 'rb') as f:
    macaroon_bytes = f.read()
    macaroon = codecs.encode(macaroon_bytes, 'hex')
```
With the connection made, we can now make our API calls. Here's a simple RPC.
```python
req = lani.TestRequest()
print(lani_stub.TestCommand(req, metadata=[('macaroon', macaroon)]))
```
Here's a response streaming RPC
```python
req = plugin.PluginRequest(name="rng-plugin")
for frame in plugin_stub.Subscribe(req, metadata=[('macaroon', macaroon)]) :
    print(frame)
```
