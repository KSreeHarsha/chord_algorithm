package triblab
import . "trib"
import "sync"
//import "log"
import "fmt"
import "time"
import "encoding/json"
import "sort"
import "math"

/*
 * TODO replace users-list with KeyValue pairs (performance)
 */

type TServer struct {
	followLock sync.Mutex
	signupLock sync.Mutex
	storage BinStorage
	userCache []string
}

const (
	USERS = "USERS"
	POSTS = "POSTS"
	FOLLOWING = "FOLLOWING"
	EMPTY_STRING = ""
)

// Utilities
func validateUsername(user string) error {
	if len(user) > MaxUsernameLen {
		return fmt.Errorf("username %q too long", user)
	}
	if !IsValidUsername(user) {
		return fmt.Errorf("invalid username %q", user)
	}
	return nil;
}

func inArray(needle string, haystack[] string) bool {
	for i := range haystack {
		if haystack[i] == needle { return true }
	}
	return false
}

func Append(slice []string, element string) []string {
	n := int(math.Max(float64(len(slice)), 0))
	list := make([]string, n + 1)
	for i := range slice {
		list[i] = slice[i]
	}
	list[n] = element
	return list
}

// Helpers
type Tribs struct {
	tribs []*Trib
}

func (d Tribs) Len() int {
	return len(d.tribs)
}

func (d Tribs) Swap(i, j int) {
	d.tribs[i], d.tribs[j] = d.tribs[j], d.tribs[i]
}

func (d Tribs) Less(i, j int) bool {
	if d.tribs[i].Clock < d.tribs[j].Clock {
		return true
	} else if d.tribs[i].Clock == d.tribs[j].Clock {
		if d.tribs[i].Time.Before(d.tribs[j].Time) {
			return true
		} else if d.tribs[i].Time.Equal(d.tribs[j].Time) {
			if d.tribs[i].User < d.tribs[j].User {
				return true
			} else if d.tribs[i].User == d.tribs[j].User {
				if d.tribs[i].Message < d.tribs[j].Message {
					return true
				}
			}
		}
	}
	return false
}

func userKey(user string) string {
	return "U" + user
}

func isEmpty(str string) bool {
	return str == EMPTY_STRING
}

// Service Methods
func (self *TServer) acquireBin(user string) Storage {
	// TODO optimize by maintaining a list of recent connections (persistent)
	return self.storage.Bin(user)
}

func (self *TServer) userList(users *List) error {
	if (len(self.userCache) < MinListUser) {
		b := self.acquireBin(USERS)
		e := b.Keys(&Pattern{ Prefix: "U", Suffix: "" }, users)
		if e != nil { return e }
		for i := range users.L {
			users.L[i] = users.L[i][1 : len(users.L[i])]
		}
		self.userCache = users.L
	}
	users.L = self.userCache
	return nil
}

func (self *TServer) userExists(user string) bool {
	var v string
	b := self.acquireBin(USERS)
	e := b.Get(userKey(user), &v)
	if e != nil {
		//log.Printf("Error while looking for user %s: %s\n", user, e)
		return true;
	}
	return !isEmpty(v)
}

func (self *TServer) getUsers() []string {
	users := new(List)
	e := self.userList(users)
	if e != nil {
		//log.Printf("Error while trying to get the list of users: %s\n", e)
		return nil
	}
	return users.L
}

// Interfaces
func (self *TServer) SignUp(user string) error {

	err := validateUsername(user)
	if err != nil { return err }

	self.signupLock.Lock()
	defer self.signupLock.Unlock()

	if self.userExists(user) {
		return fmt.Errorf("user %q already exists!", user)
	}

	var ok bool
	b := self.acquireBin(USERS)
	kv := KeyValue{ Key: userKey(user), Value: "Registered" }
	e := b.Set(&kv, &ok)

	if e == nil && !ok {
		return fmt.Errorf("failed creating new user!")
	}

	return e
}

func (self *TServer) ListUsers() ([]string, error) {
	// TODO optimize by first looking at local cache
	return self.getUsers(), nil;
}

func (self *TServer) Follow(who, whom string) error {

	self.followLock.Lock()
	defer self.followLock.Unlock()

	t, e := self.IsFollowing(who, whom)

	if e != nil { return e }
	if t != false {
		return fmt.Errorf("%s is already following %s", who, whom)
	}

	b := self.acquireBin(who)

	following, err := self.Following(who)
	if err != nil {
		return err
	}

	if len(following) >= MaxFollowing {
		return fmt.Errorf("You have already reached the limit!")
	}

	var OK bool
	e = b.ListAppend(&KeyValue{ Key: FOLLOWING, Value: whom}, &OK)
	if OK != true { return fmt.Errorf("Unable to create new follower!") }

	return e
}

func (self *TServer) IsFollowing(who, whom string) (bool, error) {

	if who == whom {
		return false, fmt.Errorf("You cannot become your follower!")
	}

	if !self.userExists(who) {
		return false, fmt.Errorf("Source user does not exist!")
	}

	if !self.userExists(whom) {
		return false, fmt.Errorf("Target user does not exist!")
	}

	users, e := self.Following(who)
	if e != nil {
		return false, e
	}

	for i := range users {
		if users[i] == whom {
			return true, nil
		}
	}

	return false, nil
}

func (self *TServer) Unfollow(who, whom string) error {

	self.followLock.Lock()
	defer self.followLock.Unlock()

	if who == whom {
		return fmt.Errorf("cannot unfollow oneself!")
	}

	b, e := self.IsFollowing(who, whom)

	if e != nil {
		return e
	}

	if b != true {
		return fmt.Errorf("user %q is not following %q", who, whom)
	}

	bin := self.acquireBin(who)
	var n int
	e = bin.ListRemove(&KeyValue{ Key: FOLLOWING, Value: whom }, &n)

	if n != 1 {
		return fmt.Errorf("expecting to see 1 while %s doing unflow for %s, %d seen", who, whom, n)
	}

	return e
}

func (self *TServer) Following(who string) ([]string, error) {

	if self.userExists(who) != true {
		return nil, fmt.Errorf("user %s does not exist!", who)
	}

	b := self.acquireBin(who)
	var list List
	e := b.ListGet(FOLLOWING, &list)

	if e != nil {
		return make([]string, 0), e
	}

	return list.L, nil
}

func (self *TServer) Post(user, post string, c uint64) error {

	if len(post) > MaxTribLen {
		return fmt.Errorf("trib too long")
	}

	if !self.userExists(user) {
		return fmt.Errorf("user does not exist!")
	}

	t := new(Trib)
	t.Time = time.Now()
	t.User = user
	t.Message = post
	t.Clock = c;

	j, e := json.Marshal(t)
	if e != nil {
		return fmt.Errorf("error while marshaling the post!")
	}

	var OK bool
	b := self.acquireBin(user)
	e = b.ListAppend(&KeyValue{ Key:POSTS, Value: string(j) }, &OK)

	if OK != true {
		return fmt.Errorf("failed while trying to save the post!")
	}

	return e
}

func (self *TServer) Home(user string) ([]*Trib, error) {

	if !self.userExists(user) {
		return make([]*Trib, 0), fmt.Errorf("user does not exist!")
	}

	var list List
	b := self.acquireBin(user)
	e := b.ListGet(FOLLOWING, &list)

	if e != nil {
		return make([]*Trib, 0), e
	}

	list.L = Append(list.L, user)

	// TODO make parallel calls to enhance performance here (optimization)
	var tc int = 0
	fts := make([]Tribs, len(list.L))
	for i := range list.L {
		t, err := self.Tribs(list.L[i])
		if err != nil {
			return make([]*Trib, 0), err
		}
		fts[i].tribs = t
		tc += len(t)
	}

	tribs := make([]*Trib, tc)
	tc = 0
	for i := range fts {
		for j := range fts[i].tribs {
			tribs[tc] = fts[i].tribs[j]
			tc++
		}
	}

	var temp Tribs
	temp.tribs = tribs
	sort.Sort(temp)

	if len(temp.tribs) > MaxTribFetch {
		temp.tribs = temp.tribs[0:MaxTribFetch - 1]
	}

	return temp.tribs, nil
}

func (self *TServer) Tribs(user string) ([]*Trib, error) {

	if !self.userExists(user) {
		return make([]*Trib, 0), fmt.Errorf("user does not exist!")
	}

	var list List
	b := self.acquireBin(user)
	e := b.ListGet(POSTS, &list)

	if e != nil {
		return make([]*Trib, 0), e
	}

	if len(list.L) > MaxTribFetch {
		list.L = list.L[0:MaxTribFetch - 1]
	}

	tribs := make([]*Trib, len(list.L))
	for i := range list.L {
		var t Trib
		e = json.Unmarshal([]byte(list.L[i]), &t)
		if e != nil {
			return make([]*Trib, 0), e
		}
		tribs[i] = &t
	}

	ts := new(Tribs)
	ts.tribs = tribs
	sort.Sort(ts)

	return ts.tribs, nil
}

var _ Server = new(TServer)
