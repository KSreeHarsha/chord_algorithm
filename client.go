package triblab
import . "trib"
import "net/rpc"
//import "fmt"
import "strings"

type client struct {
	addr string
	ns string
}

func (self *client) acquireConnection() (*rpc.Client, error) {
	return rpc.Dial("tcp", self.addr);
}

func (self *client) makeNS(key string) string {
	if (self.ns == "") {
		return key
	} else {
		return self.ns + "::" + key
	}
}

func removeNS(entry string) string {
	t := strings.Split(entry, "::")
	if len(t) == 2 {
		return t[1]
	} else {
		return t[0]
	}
}

func extractNS(entry string) string {
	t := strings.Split(entry, "::")
	if len(t) == 2 {
		return t[0]
	} else {
		return t[0] // EMPTY_STRING?!
	}
}

/*
 * Implementing KeyString
 */
func (self *client) Get(key string, value *string) error {
	c, err := self.acquireConnection();
	if err == nil {
		err = c.Call("Storage.Get", self.makeNS(key), value);
		c.Close()
	}
	return err;
}

func (self *client) Set(kv *KeyValue, succ *bool) error {
	c, err := self.acquireConnection();
	if err == nil {
		kv2 := KeyValue{ Key: self.makeNS(kv.Key), Value: kv.Value }
		err = c.Call("Storage.Set", &kv2, succ);
		c.Close()
	}
	return err;
}

func (self *client) Keys(p *Pattern, list *List) error {
	c, err := self.acquireConnection();

	if list == nil {
		list = new(List);
	}

	if err == nil {
		p2 := Pattern{ Prefix: self.makeNS(p.Prefix), Suffix: p.Suffix }
		err = c.Call("Storage.Keys", &p2, list);
		c.Close()
	}

	if list == nil {
		list = new(List)
	}

	if list.L == nil {
		list.L = make([]string, 0)
	}

	return err;
}

/*
 * Implementing KeyList
 */
func (self *client) ListGet(key string, list *List) error {
	c, err := self.acquireConnection();

	list.L = nil
	if err == nil {
		err = c.Call("Storage.ListGet", self.makeNS(key), list);
		c.Close()
	}

	if list.L == nil {
		list.L = make([]string, 0)
	}

	return err;
}

func (self *client) ListAppend(kv *KeyValue, succ *bool) error {
	c, err := self.acquireConnection();
	if err == nil {
		kv2 := KeyValue{ Key: self.makeNS(kv.Key), Value: kv.Value }
		err = c.Call("Storage.ListAppend", &kv2, succ);
		c.Close()
	}
	return err;
}

func (self *client) ListRemove(kv *KeyValue, n *int) error {
	c, err := self.acquireConnection();
	if err == nil {
		kv2 := KeyValue{ Key: self.makeNS(kv.Key), Value: kv.Value }
		err = c.Call("Storage.ListRemove", &kv2, n);
		c.Close()
	}
	return err;
}

func (self *client) ListKeys(p *Pattern, list *List) error {
	c, err := self.acquireConnection();

	list.L = nil
	if err == nil {
		p2 := Pattern{ Prefix: self.makeNS(p.Prefix), Suffix: p.Suffix }
		err = c.Call("Storage.ListKeys", &p2, list);
		c.Close()
	}

	if list.L == nil {
		list.L = make([]string, 0)
	}

	return err;
}

/*
 * Implementing Storage
 */
func (self *client) Clock(atLeast uint64, ret *uint64) error {
	c, err := self.acquireConnection();
	if err == nil {
		err = c.Call("Storage.Clock", atLeast, ret);
		c.Close()
	}
	return err;
}

var _ Storage = new(client)
