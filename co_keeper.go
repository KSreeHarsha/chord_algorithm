package triblab
import . "trib"
import "time"
import "fmt"
//import "log"
import "trib/store"

const (
	CHORD_STORE_KEY = "CHORD"
	CK_STATE_KEY	= "STATE"
	STATE_SECONDARY = "SECONDARY"
	STATE_PRIMARY	= "PRIMARY"
)

type CoKeeper struct {
	config		*KeeperConfig
	_store		Storage
	_conns		[]*client
	_myAddress	string
}

func (self *CoKeeper) init() {

	self._myAddress = self.config.Addrs[self.config.This]

	// connection handlers
	self._conns = make([]*client, len(self.config.Addrs))
	for i := range self.config.Addrs {
		if self.config.Addrs[i] == self._myAddress { continue }
		self._conns[i] = &client{ addr: self.config.Addrs[i] }
	}

	// RPC server config
	self._store = store.NewStorage()
	back := &BackConfig{
		Addr:  self._myAddress,
		Store: self._store,
		Ready: make(chan bool, 1),
	}

	// start with secondary
	self.updateMyState(STATE_SECONDARY)

	// run RPC server
	go ServeBack(back)
}

func (self *CoKeeper) updateMyState(state string) error {
	var success bool
	return self._store.Set(&KeyValue{ Key: CK_STATE_KEY, Value: state }, &success)
}

func (self *CoKeeper) run(ch *chan bool) {

	var tempChordObj, maxChordObj, tempState string

	for {

		var primaryObserved bool = false
		var anyOtherAlive bool = false
		var maxObservedAddress string = EMPTY_STRING

		// rest for a while
		time.Sleep(1 * time.Second)

		// Pull everybody
		for i := range self.config.Addrs {

			if self.config.Addrs[i] == self._myAddress { continue }

			// pull chord
			err := self._conns[i].Get(CHORD_STORE_KEY, &tempChordObj)
			if err != nil { // probably the other keeper is down
				//log.Printf("unable to read CHORD from %s", self._conns[i].addr)
				continue
			}

			err = self._conns[i].Get(CK_STATE_KEY, &tempState)
			if err != nil { // probably the other keeper is down
				//log.Printf("unable to read STATE from %s", self._conns[i].addr)
				continue
			}

			anyOtherAlive = true
			if tempState == STATE_PRIMARY {
				primaryObserved = true
				maxChordObj = tempChordObj
			}

			//log.Printf("Got CHORD from %s", self._conns[i].addr)

			if !primaryObserved && self.config.Addrs[i] > maxObservedAddress {
				maxObservedAddress = self.config.Addrs[i]
				maxChordObj = tempChordObj
			}
		}

		// decide about the future!
		// NOTE removed (maxObservedAddress != EMPTY_STRING) from second part as it can be infered from anyOtherAlive
		if !anyOtherAlive || (!primaryObserved && self._myAddress > maxObservedAddress) {
			if self.updateMyState(STATE_PRIMARY) == nil { break }
		} else {
			var success bool
			kv := KeyValue{ Key: CHORD_STORE_KEY, Value: maxChordObj }
			er := self._store.Set(&kv, &success)
			if er != nil { /*log.Printf("unable to update local CHORD: %s", er)*/ }
			if er == nil && !success { /*log.Printf("unable to update local CHORD!")*/ }
		}
	}

	//log.Printf("<%s> is the PRIMARY from now on ...", self._myAddress)
	*ch <- true // I am the PRIMARY from now on
}

/*
 * Call this method inside PRIMARY keeper for every
 * update that you make on your local Chord
 */
func (self *CoKeeper) UpdateChord(chord []byte) error {
	var success bool
	kv := KeyValue{ Key: CHORD_STORE_KEY, Value: string(chord) }
	er := self._store.Set(&kv, &success)
	if !success {
		return fmt.Errorf("unable to read chord")
	}
	return er
}

/*
 * Call this method immediately after becoming primary
 * and inside the main function of keeper
 */
func (self *CoKeeper) GetMostUpdatedChord() ([]byte, error) {
	var value string
	err := self._store.Get(CHORD_STORE_KEY, &value)
	if err == nil && value == EMPTY_STRING {
		err = fmt.Errorf("empty chord")
	}
	return []byte(value), err
}
