package triblab
import . "trib"
import "encoding/json"
//import "log"

func Sync(backend string, replicas []string, chord *Chord, c *chan bool) {
	e := _sync(backend, replicas, chord)
	*c<-(e == nil)
}

/*
 * TODO test basic operation (without any failures)
 * TODO handle keeper failures (log buffer at BackEnds)
 * TODO GC before return (close connections)
 * TODO we might be able to apply cached opLog on Primary Failure
 */
func _sync(backend string, replicas []string, chord *Chord) error {

	// Backend RPC handler
	be := &client{ addr: backend }
	rs := make([]*client, len(replicas))
	for i := range replicas {
		rs[i] = &client{ addr: replicas[i] }
	}

	prev_ip, e := chord.Prev_node_ip(backend)
	prev := &client{ addr: prev_ip }

	// 1 - Retrieve OpLog
	var opLog List
	e = be.ListGet(LOG_KEY, &opLog)
	if e != nil { return e }

	for i := range opLog.L {

		// 2 - Decode OpLog entry
		var op OpLogEntry
		e = json.Unmarshal([]byte(opLog.L[i]), &op)
		if e == nil {

			/*
			 * if the current node is the owner, I just need to replicate data to the next and next_next node
			 * otherwise, the previous node should be the owner, so I should write it to the previous and next node
			 */

			// 2.1 - I always have to write the data to the next node (it's either 1st or 2nd replica)
			e = _doWhatISay(rs[0], &op)
			if e != nil {
				//log.Printf("error while replicating [0]: %s", e)
				return e
			}

			// 2.2 - is it mine?
			owner, err := chord.getIPbyBinName(extractNS(op.Data.Key))
			if err != nil {
				//log.Printf("error getting bin name: %s", err)
			}
			//log.Printf("bin for %s = %s", op.Data.Key, owner)

			if owner == backend {

				// 2.3.1 - create the other replica
				/*
				 * avoid storing data on the same node twice (already taken care of by replication)
				 */
				//if (replicas[1] != backend) {
					e = _doWhatISay(rs[1], &op)
					if e != nil {
						//log.Printf("error while replicating [1]: %s", e)
						return e
					}
				//}

			} else if owner == prev_ip {
				//log.Printf("WARNING: you wrote the data to the successor!")
				// 2.3.2 - Write to the real owner
				e = _doWhatISay(prev, &op)
				if e != nil {
					//log.Printf("error while replicating [prev]: %s", e)
					return e
				}
			} else {
				//log.Printf("DANGER: you wrote the data to some random place!!!!!!")
			}
		}

		// 5 - Remove log entry
		var n int
		e = be.ListRemove(&KeyValue{ Key: LOG_KEY, Value: opLog.L[i] }, &n)
		if e != nil { return e }
	}
	return nil
}

/*
 * TODO do we need to have &b and &n outside this method?
 */
func _doWhatISay(c *client, o *OpLogEntry) error {

	var b bool
	var n int

	switch {
	case o.OpCode == OP_SET:
		//log.Printf("Replicating SET to %s", c.addr)
		return c.Set(&o.Data, &b)
	case o.OpCode == OP_LIST_APPEND:
		//log.Printf("Replicating LIST_APPEND to %s", c.addr)
		return c.ListAppend(&o.Data, &b)
	case o.OpCode == OP_LIST_REMOVE:
		//log.Printf("Replicating LIST_REMOVE to %s", c.addr)
		return c.ListRemove(&o.Data, &n)
	}

	return nil;
}
