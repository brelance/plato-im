# Gateway
## Task Decomposition
1. infra:TCP reason
2. infra: workpool
3. infra: epool -> epoller 
4. listeners listening clients' requests
5. infra: protobuf(cmd) / fork prpc
7. gateway RPC client(state client) -> state RPC server (Forward messages to state server)
8. implement forward messges
9. IM protocol: CMD
10. state RPC client -> gateway RPC server (Forward messages from state server to clients)