package triblab
import . "trib"

const (
	LOG_KEY = "OPLOG"

	OP_SET = 0 // opCode for Get()
	OP_LIST_APPEND = 1 // opCode for ListAppend()
	OP_LIST_REMOVE = 2 // opCode for ListRemove()
)

/*
 * Our OpLog will be a list of encoded OpLogEntries, ordered by Primary Time
 */
type OpLogEntry struct {
	OpCode uint `json:"OpCode"` // OP_SET | OP_LIST_APPEND | OP_LIST_REMOVE
	Data KeyValue `json:"Data"`
}

// TODO helper functions (e.g. encoder and decoder)
