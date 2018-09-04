# Network Frames
## Server
Receives events from client, applies events, responds with state deltas.
Keeps a buffer of deltas/events per frame, and client can request to reset their state to specific frames.
Events from clients are tied to specific frames, and as long as the event does not predate the buffer, it can be applied.
Clients can confirm that they've reached a particular frame, and the server will send the delta from the last confirmed frame to the current.
## Client
Should receive frame/delta combination and do their best to update their own state to match the world based on the frame and delta.
The client should also record local events and send them on to the server.

