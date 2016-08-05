/*
library for connecting go applications into i2p with relative ease

implements net.Listener, net.Conn, net.Addr, net.Dial, net.PacketConn that goes over i2p

    package main

    import (
      "github.com/majestrate/i2p-tools/lib/i2p"
      "fmt"
      "net"
      "net/http"
      "path/filepath"
    )

    // see i2p.StreamSession interface for more usage

    func main() {
       var err error
       var sess i2p.StreamSession
       // connect to an i2p router
       // you can pass in "" to generate a transient session that doesn't save the destination keys
       sess, err = i2p.NewStreamSessionEasy("127.0.0.1:7656", filepath.Join("some", "path", "to", "privatekey.txt"))
       if err != nil {
         fmt.Printf("failed to open connection to i2p router: %s", err.Error())
         return
       }
       // close our connection to i2p when done
       defer sess.Close()

       // i2p.Session implements net.Listener
       // we can pass it to http.Serve to serve an http server via i2p
       fmt.Printf("http server going up at http://%s/", sess.B32())
       err = http.Serve(sess, nil)
    }


*/
package i2p
