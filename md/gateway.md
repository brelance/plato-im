# Gateway
## Task Decomposition
1. infra:TCP reason
2. infra: workpool
3. infra: epool -> epoller 
4. listeners listening clients' requests
5. infra: protobuf(cmd) / fork prpc
6. gateway RPC client -> state RPC server (Forward messages to state server)
7. implement forward messges
8. IM protocol: CMD
9. state RPC client -> gateway RPC server (Forward messages from state server to clients)