# master-s-sock
sock.NewHub() // Returns Hub object
hub_object.Run() // Starts Hub
sock.ServeWs(hub_object, w, r)// Use websocket protacol on webpage
hub_object.Input // Channel that returns last string from websocket on client
