package zmq4

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const CURVE_ALLOW_ANY = "*"

var (
	auth_lock    sync.Mutex
	auth_handler *Socket
	auth_quit    = make(chan interface{})

	auth_init            = false
	auth_verbose         = false

	auth_allow = make(map[string]bool)
	auth_deny  = make(map[string]bool)

	auth_users = make(map[string]map[string]string)

	auth_pubkeys = make(map[string]map[string]bool)
)

func auth_do_handler(state State) error {
	msg, err := auth_handler.RecvMessage(0)
	if err != nil {
		if auth_verbose {
			fmt.Println("AUTH: Terminating")
		}
		return errors.New("TERM")
	}

	version := msg[0]
	if version != "1.0" {
		panic("AUTH: version != 1.0")
	}

	sequence := msg[1]
	domain := msg[2]
	address := msg[3]
	//identity := msg[4]
	mechanism := msg[5]

	username := ""
	password := ""
	//client_key := ""
	if mechanism == "PLAIN" {
		username = msg[6]
		password = msg[7]
	} else if mechanism == "CURVE" {
		s := msg[6]
		if len(s) != 32 {
			panic("AUTH: len(client_key) != 32")
		}
		//client_key = Z85encode(s)
	}

	allowed := false
	denied := false

	if len(auth_allow) > 0 {
		if auth_allow[address] {
			allowed = true
			if auth_verbose {
				fmt.Printf("AUTH: PASSED (whitelist) address=%s\n", address)
			}
		} else {
			denied = true
			if auth_verbose {
				fmt.Printf("AUTH: DENIED (not in whitelist) address=%s\n", address)
			}
		}
	} else if len(auth_deny) > 0 {
		if auth_deny[address] {
			denied = true
			if auth_verbose {
				fmt.Printf("AUTH: DENIED (blacklist) address=%s\n", address)
			}
		} else {
			allowed = true
			if auth_verbose {
				fmt.Printf("AUTH: PASSED (not in blacklist) address=%s\n", address)
			}
		}
	}

	//  Mechanism-specific checks
	if !denied {
		if mechanism == "NULL" && !allowed {
			//  For NULL, we allow if the address wasn't blacklisted
			if auth_verbose {
				fmt.Printf("AUTH: ALLOWED (NULL)\n")
			}
			allowed = true
		} else if mechanism == "PLAIN" {
			//  For PLAIN, even a whitelisted address must authenticate
			allowed = authenticate_plain(domain, username, password)
		} else if mechanism == "CURVE" {
			//  For CURVE, even a whitelisted address must authenticate
			allowed = authenticate_curve()
		}
	}
	if allowed {
		auth_handler.SendMessage(version, sequence, "200", "OK", "anonymous", "")
	} else {
		auth_handler.SendMessage(version, sequence, "400", "NO ACCESS", "", "")
	}
	return nil
}

func authenticate_plain(domain, username, password string) bool {
	for _, dom := range []string{domain, "*"} {
		if m, ok := auth_users[dom]; ok {
			if m[username] == password {
				if auth_verbose {
					fmt.Printf("AUTH: ALLOWED (PLAIN) domain=%s username=%s password=%s\n", dom, username, password)
				}
				return true
			}
		}
	}
	if auth_verbose {
		fmt.Printf("AUTH: DENIED (PLAIN) domain=%s username=%s password=%s\n", domain, username, password)
	}
	return false
}

func authenticate_curve() bool {
	return true
}

func auth_do_quit(i interface{}) error {
	if auth_verbose {
		fmt.Println("AUTH: Quit")
	}
	return errors.New("Quit")
}

func AuthStart() (err error) {
	auth_lock.Lock()
	defer auth_lock.Unlock()

	if auth_init {
		if auth_verbose {
			fmt.Println("AUTH: Already running")
		}
		return errors.New("Auth is already running")
	}
	auth_init = true
	if auth_verbose {
		fmt.Println("AUTH: Starting")
	}

	auth_handler, err = NewSocket(REP)
	if err != nil {
		return
	}
	err = auth_handler.Bind("inproc://zeromq.zap.01")
	if err != nil {
		return
	}

	reactor := NewReactor()
	reactor.AddSocket(auth_handler, POLLIN, auth_do_handler)
	reactor.AddChannel(auth_quit, 0, auth_do_quit)
	go func() {
		reactor.Run(100 * time.Millisecond)
		err := auth_handler.Close()
		if auth_verbose && err != nil {
			fmt.Println("AUTH:", err)
		}
		if auth_verbose {
			fmt.Println("AUTH: Finished")
		}
		auth_init = false
	}()
	if auth_verbose {
		fmt.Println("AUTH: Started")
	}
	return
}

func AuthStop() {
	if auth_verbose {
		fmt.Println("AUTH: Stopping")
	}
	auth_quit <- true
	time.Sleep(100 * time.Millisecond)
}

//  Allow (whitelist) some IP addresses. For NULL, all clients from these
//  addresses will be accepted. For PLAIN and CURVE, they will be allowed to
//  continue with authentication. You can call this method multiple times
//  to whitelist multiple IP addresses. If you whitelist a single address,
//  any non-whitelisted addresses are treated as blacklisted.
func AuthAllow(addresses ...string) {
	for _, address := range addresses {
		auth_allow[address] = true
	}
}

//  Deny (blacklist) some IP addresses. For all security mechanisms, this
//  rejects the connection without any further authentication. Use either a
//  whitelist, or a blacklist, not not both. If you define both a whitelist
//  and a blacklist, only the whitelist takes effect.
func AuthDeny(addresses ...string) {
	for _, address := range addresses {
		auth_deny[address] = true
	}
}

//  Configure PLAIN authentication for a given domain.
//  The `users` is a map from username to passwords.
//  Set `domain` to "*" to apply to all domains.
func AuthConfigurePlain(domain string, users map[string]string) {
	if _, ok := auth_users[domain]; !ok {
		auth_users[domain] = make(map[string]string)
	}
	for key := range users {
		auth_users[domain][key] = users[key]
	}
}

//  Configure CURVE authentication for a given domain. CURVE authentication
//  uses a directory that holds all public client certificates, i.e. their
//  public keys. The certificates must be in zcert_save () format. The
//  location is treated as a printf format. To cover all domains, use "*".
//  You can add and remove certificates in that directory at any time.
//  To allow all client keys without checking, specify CURVE_ALLOW_ANY
//  for the location.

func AuthConfigureCurve(domain string, pubkeys ...string) {
	if _, ok := auth_pubkeys[domain]; !ok {
		auth_pubkeys[domain] = make(map[string]bool)
	}
	for key := range pubkeys {
		auth_pubkeys[domain][key] = true
	}
}

// Enable verbose tracing of commands and activity
func AuthSetVerbose(verbose bool) {
	auth_verbose = verbose
}
