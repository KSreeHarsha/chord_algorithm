package triblab
import "fmt"
import "log"
import "encoding/json"
import "time"
import . "trib"
import "trib/store"

const (
	GARBAGE_KVP = 90
	GARBAGE_LIST = 99
)

type Garbage struct {
	Key		string	`json:"Key"`
	Type	uint	`json:"Type"`
}

func (self *Garbage) String() string {
	var _type string
	if self.Type == GARBAGE_LIST { _type = "List" } else { _type = "KVP" }
	return fmt.Sprintf("Garbage %s of type %s", self.Key, _type)
}

type GarbageCollector struct {
	trashCan Storage
	chord *Chord
}

func (self *GarbageCollector) removeCan(backend string) {
	var n int
	var items List

	err := self.trashCan.ListGet(backend, &items)
	if err != nil {
		log.Printf("removeCan(%s) failed: %s", backend, err)
		return
	}

	for i := range items.L {
		kv := &KeyValue{ Key: backend, Value: items.L[i] }
		err = self.trashCan.ListRemove(kv, &n)
		if err != nil {
			log.Printf("removeCan(%s) warning: %s", backend, err)
			continue
		}
	}
}

func (self *GarbageCollector) clean(backend string, ch *chan bool) {
	var n int
	var succ bool
	var items List

	err := self.trashCan.ListGet(backend, &items)
	if err != nil {
		log.Printf("clean(%s) failed[1]: %s", backend, err)
		*ch <- false; return
	}

	c := &client{ addr: backend }

	for i := range items.L {
		var gb Garbage
		json.Unmarshal([]byte(items.L[i]), &gb)

		log.Printf("cleaning %s for %s", backend, gb)

		if gb.Type == GARBAGE_KVP {
			err = c.Set(&KeyValue{ Key: gb.Key, Value: "" }, &succ)
			if err != nil {
				log.Printf("clean(%s) failed[2]: %s", backend, err)
				*ch <- false
				return
			}
		} else if gb.Type == GARBAGE_LIST {
			var t List
			err = c.ListGet(gb.Key, &t)
			if err != nil {
				log.Printf("clean(%s) failed[3]: %s", backend, err)
				*ch <- false
				return
			}
			for j := range t.L {
				err = c.ListRemove(&KeyValue{ Key: gb.Key, Value: t.L[j] }, &n)
				if err != nil {
					log.Printf("clean(%s) failed[4]: %s", backend, err)
					*ch <- false
					return
				}
			}
		} else {
			log.Printf("clean(%s) invalid operation!", backend)
		}

		kv := &KeyValue{ Key: backend, Value: items.L[i] }
		err = self.trashCan.ListRemove(kv, &n)
		if err != nil {
			log.Printf("clean(%s) failed[5]: %s", backend, err)
			*ch <- false; return
		}
	}

	*ch <- true
}

func (self *GarbageCollector) run() {

	self.trashCan = store.NewStorage()

	for {
		// sleep
		time.Sleep(5 * time.Second)
		//log.Printf("GC woke up ...")

		var keys List
		p := Pattern{ Prefix: "", Suffix: "" }
		err := self.trashCan.ListKeys(&p, &keys)
		if err != nil {
			//log.Printf("GC error [1]: %s", err)
			continue
		}

		// channel
		ch := make(chan bool, len(keys.L))

		for i := range keys.L {
			bin, e := self.chord.getIPbyBinName(keys.L[i])
			if e != nil || bin != keys.L[i] {
				ch <- false
				self.removeCan(keys.L[i]) // already dead!
				continue
			}
			go self.clean(keys.L[i], &ch)
		}

		// wait for workers
		for i := 0; i < len(keys.L); i++ { <-ch }
		//log.Printf("GC went to sleep ...")
	}
}

func (self *GarbageCollector) mark(backend string, gb *Garbage) {
	var succ bool
	//log.Printf("%s - [%s]", gb, backend)
	jso, _ := json.Marshal(gb)

	// TODO avoid duplication
	self.trashCan.ListAppend(&KeyValue{ Key: backend, Value: string(jso) }, &succ)
}
