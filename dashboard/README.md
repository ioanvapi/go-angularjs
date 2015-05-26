




## Close the browser window

1. ReadPings() from Connection get error (err = EOF) when ws.ReadMessage(). That's because browser websocket connection is closed.
2. We terminate ReadPings() and signal to 'unregister' the connection (and close ws on server side).
3. Hub receives the unregister for connection and delete the connection from internal map and close its communication channel 'send'.
4. StartWriting() goroutine detects the send channel is closed and stops itself then signal to 'unregister' the connection (and close ws on server side). 
5. Hub receives again the unregister for connection but find its already unregistered.
 
 