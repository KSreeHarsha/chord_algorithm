package triblab

import . "trib"

var LogEnabled bool = false

// Creates an RPC client that connects to addr.
func NewClient(addr string) Storage {
	//return &client{ addr: addr }
	cli := &OpLogClient{ addr: addr }
	cli.init()
	return cli
}

// Serve as a backend based on the given configuration
func ServeBack(b *BackConfig) error {
	s := server{};
	return s.init(b);
}
