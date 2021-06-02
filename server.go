package triblab

import . "trib"
import "fmt"
import "net"
import "net/rpc"
import "net/http"

type server struct { }

func (self *server) init(b *BackConfig) error {

	rpcServer := rpc.NewServer();
	err := rpcServer.Register(b.Store);
	if err != nil {
		b.Ready <- false;
		return err;
	}

	http.Handle("/rpc" + b.Addr, rpcServer);

	l, e := net.Listen("tcp", b.Addr);
	if e != nil {
		b.Ready <- false;
		return err;
	}

	if b.Ready != nil {
		b.Ready <- true;
	}

	for {
		if conn, err := l.Accept(); err != nil {
			fmt.Printf("Failed accepting connection: " + err.Error() + "\n");
		} else {
			go rpcServer.ServeConn(conn)
		}
	}

	return nil;
}


