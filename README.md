# yameee
Yet another messenger, but with some differences: 
 - no any data is stored on server
 - no permanent accounts
 - just temporary secure pipes between peers
 
WebSocket as main communication protocol. Messages are encrypted with AES and Ed25519 signature. 
As a client - simple JS window. But all client-side calcullations are developed on Go and implemented to JS as WebAssembly

WORK IS STILL IN PROGRESS
Details will be added soon

Disclaimer
Messenger is developed for fun only and testing all technologies (WS, WA etc)
