#!/usr/bin/env bash

CLIENT_ID=web_test_client
MAC_ADDRESS=$(cat testdata/01-initial-request.json | jq -r '.mac_address')

# [DEVICE] First contact request from device to discover websocket URL for it to connect to
xh -v POST http://127.0.0.1:8090/xiaozhi/ota @testdata/01-initial-request.json "client-id:${CLIENT_ID}" "device-id:${MAC_ADDRESS}"

# 2. [WEBSOCKET_SERVER]
#xh -v localhost:8090/xiaozhi/config/agent-models "clientId=${CLIENT_ID}" "macAddress=${MAC_ADDRESS}" "selectedModule[ASR]=ASR_FunASR" "selectedModule[VAD]=VAD_SileroVAD"