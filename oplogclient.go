package triblab
import . "trib"
import "encoding/json"
import "fmt"

type OpLogClient struct {
	addr string
	ns string
	cli *client
	_log *client
}

func (self *OpLogClient) makeNS(key string) string {
	if (self.ns == "") {
		return key
	} else {
		return self.ns + "::" + key
	}
}

func (self *OpLogClient) init() {
	self.cli = &client{ addr: self.addr, ns: self.ns }
	self._log = &client{ addr: self.addr }
}

/*
 * Implementing KeyString
 */

func (self *OpLogClient) Get(key string, value *string) error {
	return self.cli.Get(key,value);
}

func (self *OpLogClient) log(opCode uint, kv *KeyValue) error {
	var succ bool

	op := &OpLogEntry{ OpCode: opCode, Data: *kv }
	jsonObj, e := json.Marshal(op)
	if e != nil {
		return fmt.Errorf("Error while marshaling the OP Code")
	}

	OPkv := KeyValue{ Key:LOG_KEY , Value: string(jsonObj) }
	e = self._log.ListAppend(&OPkv, &succ)
	if e != nil { return e }
	if !succ { return fmt.Errorf("error while appending to the log") }
	return nil
}

func (self *OpLogClient) Set(kv *KeyValue, succ *bool) error {

	kv2 := KeyValue{ Key: self.makeNS(kv.Key), Value: kv.Value }
	e := self.log(OP_SET, &kv2)
	if e != nil { return e }
	e = self.cli.Set(kv, succ)
	return e
}

func (self *OpLogClient) Keys(p *Pattern, list *List) error {
	e := self.cli.Keys(p,list);
	for i := range list.L {
		list.L[i] = removeNS(list.L[i])
	}
	return e
}

func (self *OpLogClient) ListGet(key string, list *List) error {
	return self.cli.ListGet(key,list);
}

func (self *OpLogClient) ListAppend(kv *KeyValue, succ *bool) error {

	kv2 := KeyValue{ Key: self.makeNS(kv.Key), Value: kv.Value }
	e := self.log(OP_LIST_APPEND, &kv2)
	if e != nil { return e }
	e = self.cli.ListAppend(kv, succ)
	return e
}

func (self *OpLogClient) ListRemove(kv *KeyValue, n *int) error {

	kv2 := KeyValue{ Key: self.makeNS(kv.Key), Value: kv.Value }
	e := self.log(OP_LIST_REMOVE, &kv2)
	if e != nil { return e }
	e = self.cli.ListRemove(kv, n)
	return e
}

func (self *OpLogClient) ListKeys(p *Pattern, list *List) error {
	e := self.cli.ListKeys(p,list);
	if e != nil { return e }

	// removing OpLog from ListKeys
	var opLogExists bool = false
	for i := range list.L {
		if list.L[i] == LOG_KEY {
			opLogExists = true
		} else {
			list.L[i] = removeNS(list.L[i])
		}
	}

	if !opLogExists { return nil }

	var list2 List
	list2.L = make([]string, len(list.L) - 1)
	var j uint = 0
	for i := range list.L {
		if (list.L[i] != LOG_KEY) {
			list2.L[j] = list.L[i]
			j++
		}
	}
	list.L = list2.L
	return nil
}

/*
 * Implementing Storage
 */
func (self *OpLogClient) Clock(atLeast uint64, ret *uint64) error {
	return self.cli.Clock(atLeast,ret)
}

var _ Storage = new(OpLogClient)
