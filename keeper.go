package triblab
import (
	. "trib"
	"time"
	"fmt"
//	"log"
	"errors"
    "encoding/json"
)

const (
	ALIVE = 1
	SILENT = 2
	DEAD = 3
)

const MAXHASHVAL = 4294967296 - 1
//TODO-isthis correct MAXHASHVAL?

type worker struct {
	address		string
	lastAck		uint64
	status		int
	silentDur	int
	handler		*client
}

func (w worker) String() string {
	return fmt.Sprintf("%s (%d, %d)", w.address, w.lastAck, w.status)
}

type keeper struct {
	config		*KeeperConfig
	workers		[]worker
}

//Maintain Node and Ring for the Chord Ring information
type Node struct {
    Ip string //its own ip
    Prev string //prev ip
    Succ string //succ ip
    Start uint32 //start of its arc
    End uint32 //end of its arc onr ring
}

//Initially there is no node in the ring
//var ring []Node

//create ring with zero nodes
func (self *Chord) initialize() error{
    self.Ring = make([]Node, 0)
    return nil
}

func (self *Chord) printRing() error{
    for i := range self.Ring{
        fmt.Printf("%+v",self.Ring[i])
    }
    return nil
}

func (self *Chord) lookupValinRing(val uint32) (string, error){
    if len(self.Ring)==0{
        return "",fmt.Errorf("ring is of size 0 nodes. Cannot find a node.")
    }
    if len(self.Ring)==1{
        return self.Ring[0].Ip,nil
    }
    if len(self.Ring)>1{
        for j:=0;j<len(self.Ring);j++{
            //Normal case, whent he start value is less than the end value

//log.Print(val, self.Ring[j].Start,self.Ring[j].End, self.Ring[j].Ip)
          if self.Ring[j].Start < self.Ring[j].End {
       //       log.Print("12")
            if self.Ring[j].Start <= val && val <= self.Ring[j].End{
         //       log.Print("23")
                return self.Ring[j].Ip, nil
            }
        }
            //else check for the "0" key jumpingmp case i.e. the val is between end and start
            if (self.Ring[j].Start > self.Ring[j].End){
   //             log.Print("11")
                if((self.Ring[j].End >= val && val >= 0) || (self.Ring[j].Start <= val)){
     //               log.Print("22")
                    return self.Ring[j].Ip, nil
                }
            }
        }
    }
    //log.Print(":O")
    return "", fmt.Errorf("should never reach here. Check function")
}

func (self *Chord) getIPbyBinName(name string) (string,error){
    val := getHash(name)
//log.Print("Bin name - ", name)
    return self.lookupValinRing(val)
}

func (self *Chord) listAllActiveNodes() ([]string, error){
    list_of_active_nodes:=make([]string,0)

    for i:=0;i<len(self.Ring);i++{
        list_of_active_nodes = append(list_of_active_nodes,self.Ring[i].Ip)
    }
    return list_of_active_nodes,nil
}

func (self *Chord) addNodetoRing(ip string) (string, string, error){
    var node Node
    val := getHash(ip)

/*
    _,err456:=self.getIPbyBinName("localhost:28049")
    if err456!=nil{
        log.Printf("err message is %s", err456)
    }
  */  /*//Testing - vineet
    _,err56:=self.Succ_node_ip(ip)
    if err56!=nil{
        log.Printf("error message is %s",err56 )
    }
*/

    //var next,prev string

    //Folowing values are fixed regardless of node location in ring
    node.Ip = ip
//log.Print("add1")
    if len(self.Ring)==0{
//log.Print("add2")
        node.Succ = ""
        node.Prev = ""
        node.Start = 0
        node.End = MAXHASHVAL
    } //else{
        if len(self.Ring)==1{
//log.Print("add3")
            node.Succ=self.Ring[0].Ip
            node.Prev=self.Ring[0].Ip
            node.Start=getHash(self.Ring[0].Ip)+1
            node.End=val
//log.Print("add4")

            //Fix the other node - which is already existing
            self.Ring[0].Succ=ip
            self.Ring[0].Prev=ip
            self.Ring[0].Start=val+1
            self.Ring[0].End=getHash(self.Ring[0].Ip)
//log.Print("add5")
            //So, now both the nodes are fixed. The initial node covers the "0" key
        }//else{
        //TODO-above. I think len=2 case is handled in loop. Confirm.
        //For rings with 2 nodes or more
        if len(self.Ring)>1{

//log.Print("add6 - ")
    for i:=0;i<len(self.Ring);i++{
//log.Print("add7 - ", val, self.Ring[i].start, self.Ring[i].end)
        //Normal case - when a node's start value is less than end
      if self.Ring[i].Start < self.Ring[i].End{
//log.Print("add8")
        if (val >= self.Ring[i].Start && val <= self.Ring[i].End){
//log.Print("add9")
            node.Succ = self.Ring[i].Ip
            node.Prev = self.Ring[i].Prev
            node.Start = self.Ring[i].Start
            node.End = val
            //Fix the successor node
            self.Ring[i].Prev = node.Ip
            self.Ring[i].Start = val+1
            //Fix the predecessor node later - TODO
            break
        }
      }else{
//log.Print("add10")
          //Special case - jumping over the zero key
        if self.Ring[i].Start > self.Ring[i].End{
//log.Print("add11")
            if (val <= self.Ring[i].End && val >= 0) ||  val >= self.Ring[i].Start{
//log.Print("add12")
                node.Succ=self.Ring[i].Ip
                node.Prev=self.Ring[i].Prev
                node.Start=self.Ring[i].Start
                node.End=val
                //Fixing the successor
                self.Ring[i].Prev=node.Ip
                self.Ring[i].Start=val+1
                //Fix the predecessor node later - TODO
                break
        }
    }
//log.Print("add13")
            //return "","",fmt.Errorf("some error, the node must be inserted somewhere")
    }
    }
//log.Print("add14")
    //Fix the predecessor's arc and prev value
    for j:=0;j<len(self.Ring);j++{
//log.Print("add15")
        if self.Ring[j].Ip==node.Prev{
//log.Print("add16")
            self.Ring[j].Succ = node.Ip
            break
        }
    }
}
//log.Print("add17")
    self.Ring = append(self.Ring, node) //TODO-is this how you use append
//log.Print("add18")
    return  node.Succ,node.Prev,nil
}

func (self *Chord) removeNodefromRing(ip string) (string, string, error){
//log.Print("remove1")
    if len(self.Ring)==0{
//log.Print("remove2")
        return "","",fmt.Errorf("ring already empty, cannot remove node")
    }
//log.Print("remove3")
    if len(self.Ring)==1{
//log.Print("remove4")
        if self.Ring[0].Ip==ip{
//log.Print("remove5")
            //the only node in the ring is the node we want to delete
            //create ring of size 0
            self.Ring = make ([]Node, 0)
//log.Print("remove6")
            return "","",nil
        }else{
//log.Print("remove7")
            return "","",fmt.Errorf("the only node in ring does not share ip with the node being removed. Error")
        }
    }
    var ip_used string

    if len(self.Ring)==2{
//log.Print("remove8")
        if self.Ring[0].Ip==ip{
            ip_used=self.Ring[1].Ip
//log.Print("remove9")
        }else{
        if self.Ring[1].Ip==ip{
            ip_used=self.Ring[0].Ip
//log.Print("remove10")
        }else{
//log.Print("remove11")
            return "","",fmt.Errorf("Ring has two nodes, but none of the nodes match the ip being deleted")
        }}

//log.Print("remove12 - ",)
//Modify ring to contain only one node
        self.Ring = make([]Node,1)
//log.Print("remove121")
        self.Ring[0].Succ=""
        self.Ring[0].Prev=""
//log.Print("remove122")
        self.Ring[0].Ip=ip_used
//log.Print("remove122")
        self.Ring[0].Start=0
        self.Ring[0].End= MAXHASHVAL

//log.Print("remove13")
        //above should leave only one node in the ring
        if len(self.Ring)!=1{
//log.Print("remove14")
            return "","",fmt.Errorf("ring is not of size 1. Error")
        }
//log.Print("remove15")
        return "","",nil
    }

    for i:=0;i<len(self.Ring);i++{
//log.Print("remove20")
        if ip==self.Ring[i].Ip{
			next1:=self.Ring[i].Succ
			prev1:=self.Ring[i].Prev
//log.Print("remove21")
            //we have found our node, need to modify the relevant succ and prev values in ring
            //and remove the node
            for j:=0;j<len(self.Ring);j++{
//log.Print("remove22")
            //Fix the successor node
                if self.Ring[j].Ip==self.Ring[i].Succ{
//log.Print("remove23")
                    self.Ring[j].Prev = self.Ring[i].Prev
                    self.Ring[j].Start = self.Ring[i].Start
                }
            //Fix the predecessor node
//log.Print("remove24")
                if self.Ring[j].Ip==self.Ring[i].Prev{
//log.Print("remove25")
                    self.Ring[j].Succ=self.Ring[i].Succ
                }
            }
//log.Print("remove26")
            //Remove the node

            self.Ring = append(self.Ring[:i],self.Ring[i+1:]...)
//log.Print("remove27")
            return next1,prev1,nil
        }
    }
    return "","",nil
}

/*
//TODO-only works for one keeper. The range needs to be modified when working with multiple keepers
func (self *keeper) node_status() error{
    for i:= range self.config.Backs {
        node_status[i] = false
    }
    return nil
}
*/

func (self *Chord) MarshalChord() ([]byte, error){
    return json.Marshal(self)
}

func (self *keeper) run() error {
//The section below does the wait and read/write chord data.
//The actual keeper starts at "Tag:2"
	//log.Print("starting keeper ...")
    cokeep := &CoKeeper{config:self.config}
    cokeep.init()

    keep_sync_chan := make(chan bool,1)
    go cokeep.run(&keep_sync_chan)
    <-keep_sync_chan

    bytechord1,err01:=cokeep.GetMostUpdatedChord()
    if err01!=nil{
        //log.Print("keeper.go - error in GetMostUpdatedChord")
        //return fmt.Errorf("keeper.go - Error in GetMostUpdatedChord")
    }

    //Tag:2 - BEYOND THIS LINE RUNS THE KEEPER - YOU HAVE BEEN WARNED

    //keeps a count mod 3. Everytime it is 0, we call the Clock().
    //When it is 0,1, or 2, we do the node join/crash check
    var count int = -1
    //Table maintained by Keeper for all the nodes
    //TODO-300 is static. Can we maintain this table dynamic while sharing it between keepers
    var node_status []bool = make ([]bool,300)

	// initialize
    var chord Chord
    chord.initialize()
	if err01!=fmt.Errorf("empty chord"){
		if json.Unmarshal(bytechord1, &chord) != nil {
			//log.Print("unable to unmarshal received CHORD!")
		}
	}

//log.Print(chord.Ring)
    replication := &ReplicationService{ _chord: &chord }

//log.Print(self.config.Backs)
	for i := range self.config.Backs {
		self.workers = append(self.workers, worker{
			address: self.config.Backs[i],
			lastAck: 0,
			status: ALIVE,
			handler: &client{ addr: self.config.Backs[i] },
		})
	}

	go replication.run()
//log.Print(1)
	for {
		// Heartbeat period
		time.Sleep(1 * time.Second)
        count = (count+1)%3

    //Perform the following node check operation everytime
        key := "STATUS"
        value:= "TRUE"
        keyvalue := KeyValue{Key:key,Value:value}
        succ:= false

//log.Print(2)
    for i:= range self.config.Backs {
        cur_node_status := false
        err_node_status := self.workers[i].handler.Set(&keyvalue, &succ)
        //If you are able to do above "Set" operation, then the node is updating
//log.Print(3)
        if err_node_status!=nil{
            //log.Print("not up - ",self.config.Backs[i])
		}

        if err_node_status==nil{
         if(succ==true){
                //the operation had succeeded, the node is up
                //TODO-is my understanding of succ correct
                cur_node_status=true
//log.Print(cur_node_status)
            }else{
                //node is down
                cur_node_status=false
//log.Print(cur_node_status)
            }}

//log.Print(4)
        var succ2,succ3 bool
        var next,prev string

        if node_status[i]==false{
            if cur_node_status==true{
            //new node has joined
            //Call replication service
            //modify ring - add node
			var err1 error
//log.Print(5)
//vineet
//log.Print("len of chord.Ring=", len(chord.Ring))
//log.Print(chord.Ring)
			next,prev,err1 = chord.addNodetoRing(self.config.Backs[i])
//vineet
            //log.Print(6)
//log.Print(chord.Ring)
			if err1!=nil{
				fmt.Errorf("error in adding node")
			}
            chordminisnapshot,err1:=CreateMiniChord(self.config.Backs[i],&chord)
            if err1!=nil{
                fmt.Errorf("chordminisnapshot network error. check.")
            }
//log.Print(7)
        bytechord,err10:=chord.MarshalChord()
        if err10!=nil{
            //log.Print("Error in marshaling chord")
        }
        err31:=cokeep.UpdateChord(bytechord)
        if err31!=nil{
            //log.Print("Error in sending update chord")
        }
			go replication.notifyJoin(&chordminisnapshot)
//log.Print(8)
            //add successor/previous keys on the corresponding nodes
            err2:=self.workers[i].handler.Set(&KeyValue{
                Key: "NEXT",
                Value: next},&succ2)
                if err2!=nil || succ2 !=true{
                    fmt.Errorf("Error with Set NEXT")
                }
/*
                        var value40 string
                        //Testing to see if the value is written correctly
                        err40:=self.workers[i].handler.Get("NEXT",&value40)
                        if err40!=nil{
                            log.Print("error reading NEXT value")
                        }
                        log.Print("NEXT VALUE", value40)
                        */
//log.Print(9)
            err3:=self.workers[i].handler.Set(&KeyValue{
                Key: "PREV",
                Value: prev}, &succ3)
                if err3!=nil || succ3!=true{
                    fmt.Errorf("Error with Set PREV")
                }
//log.Print(10)
/*
                        var value401 string
                        //Testing to see if the value is written correctly
                        err401:=self.workers[i].handler.Get("PREV",&value401)
                        if err401!=nil{
                            log.Print("error reading NEXT value")
                        }
                        log.Print("NEXT VALUE", value401)
*/
                //TODO-also modify these values on the other nodes
                var succ4,succ5 bool
                for j:=0;j<len(self.workers);j++{
                    if self.config.Backs[j]==prev{
                        err4:=self.workers[j].handler.Set(&KeyValue{Key:"NEXT",Value:self.config.Backs[i]},&succ4)
                        if err4!=nil || succ4!=true{
                            return fmt.Errorf("Error: with set NEXT in prev")
                        }
 /*                       var value41 string
                        //Testing to see if the value is written correctly
                        err41:=self.workers[j].handler.Get("NEXT",&value41)
                        if err41!=nil{
                            log.Print("error reading NEXT value")
                        }
                        log.Print("NEXT VALUE", value41)
   */                 }
//log.Print(11)
                    if self.config.Backs[j]==next{
                        err5:=self.workers[j].handler.Set(&KeyValue{Key:"PREV",Value:self.config.Backs[i]},&succ5)
                        if err5!=nil || succ5!=true{
                            return fmt.Errorf("Error: with set PREV in next")
                        }
    /*                    var value42 string
                        //Testing to see if the value is written correctly
                        err42:=self.workers[j].handler.Get("PREV",&value42)
                        if err42!=nil{
                            log.Print("error reading NEXT value")
                        }
                        log.Print("NEXT VALUE", value42) */
                    }
//log.Print(12)
                }
           node_status[i]=true
            }else{
                //Nothing to do
//log.Print(13)
            }
        }else {
            if cur_node_status==true{
                //Nothing to does
//log.Print(14)
            }else{
                //Node has failed
                //Call replication service
//log.Print(15)
            chordminisnapshot1,errr:=CreateMiniChord(self.config.Backs[i],&chord)
            if errr!=nil{
                fmt.Errorf("chordminisnapshot network error. check.")
            }
        bytechord12,err102:=chord.MarshalChord()
        if err102!=nil{
            //log.Print("Error in marshaling chord")
        }
        err34:=cokeep.UpdateChord(bytechord12)
        if err34!=nil{
            //log.Print("Error in sending update chord")
        }
			go replication.notifyLeave(&chordminisnapshot1)
//log.Print(16)
//vineet
//log.Print(chord.Ring)
            //Remove node - modify ring
                var err2 error
                next,prev,err2 = chord.removeNodefromRing(self.config.Backs[i])
//log.Print(17,next,prev,self.config.Backs[i])
//vineet
//log.Print(chord.Ring)
                if err2!=nil{
//log.Print(18)
                    fmt.Errorf("Error removing node.")
                }
                //TODO-Modify successor/previous keys
                //TODO for the other nodes
                var succ4,succ5 bool
                for j:=0;j<len(self.workers);j++{
                    if self.config.Backs[j]==prev{
//log.Print("194-----------------------------")
                        err4:=self.workers[j].handler.Set(&KeyValue{Key:"NEXT",Value:next},&succ4)
                        if err4!=nil || succ4!=true{
                           return fmt.Errorf("Error: with set NEXT in prev")
                        }
                    }

                    if self.config.Backs[j]==next{
//log.Print("195-----------------")
                        err5:=self.workers[j].handler.Set(&KeyValue{Key:"PREV",Value:prev},&succ5)
//log.Print(196)
                        if err5!=nil || succ5!=true{
//log.Print(197)
                            return fmt.Errorf("Error: with set PREV in next")
                        }
                    }
                }
//log.Print(198)
            node_status[i]=false
        }
    }
    //Perform the following clock sync only every 3rd second
     if count==2{
//log.Print(20)
		// Query workers
		var maxClock uint64
		for i := range self.workers {
			err := self.workers[i].handler.Clock(self.workers[i].lastAck, &self.workers[i].lastAck);
			if err == nil {
				if maxClock < self.workers[i].lastAck {
					maxClock = self.workers[i].lastAck
				}

				// Mark worker as ALIVE
				self.workers[i].status = ALIVE
			} else {
				// Print log message
				//log.Printf("Error reading from %s\n", self.workers[i])

				if self.workers[i].status != DEAD {
					// Mark worker as SILENT
					if self.workers[i].status != SILENT {
						self.workers[i].status = SILENT
						self.workers[i].silentDur = 0
					}
					self.workers[i].silentDur++
				}
			}
		}

		// Sync clocks
		for i := range self.workers {
			if self.workers[i].status == ALIVE {
				if self.workers[i].lastAck < maxClock {
					// Sync clock with maxClock
					//log.Printf("Shifting clock from %d to %d for %s\n", self.workers[i].lastAck, maxClock, self.workers[i].address)
					err := self.workers[i].handler.Clock(maxClock, &self.workers[i].lastAck)
					if err != nil {
						// Print log message
						//log.Printf("Error updating clock for %s\n", self.workers[i])
					}
				}
			} else {
				if (self.workers[i].status == SILENT && self.workers[i].silentDur == 10) {
					self.workers[i].status = DEAD
					//log.Printf("Potential failure of %s detected\n", self.workers[i].address)
				}
			}
		}
     }
	}
}
	return errors.New("Keeper terminated unexpectedly!");
}

