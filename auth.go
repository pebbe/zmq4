/*

This file implements functionality very similar to that of the xauth module in czmq.

Notable differences in here:

 - domains are supported
 - usernames/passwords are read from memory, not from file
 - public keys are read from memory, not from file
 - additional functions for configuring server or client socket with a single command

*/

package zmq4

import (
	"errors"
	"fmt"
)

const CURVE_ALLOW_ANY = "*"

var (
	auth_handler *Socket
	auth_quit    *Socket

	auth_init    = false
	auth_verbose = false

	auth_allow = make(map[string]bool)
	auth_deny  = make(map[string]bool)

	auth_users = make(map[string]map[string]string)

	auth_pubkeys = make(map[string]map[string]bool)

	auth_meta_handler = auth_meta_handler_default
)

func auth_meta_handler_default(version, request_id, domain, address, identity, mechanism string) (user_id string, metadata []byte) {
	return "", []byte{}
}

func auth_do_handler() {
	for {

		msg, err := auth_handler.RecvMessage(0)
		if err != nil {
			if auth_verbose {
				fmt.Println("AUTH: Terminating")
			}
			break
		}

		if msg[0] == "QUIT" {
			if auth_verbose {
				fmt.Println("AUTH: Quiting")
			}
			auth_handler.SendMessage("QUIT")
			break
		}

		version := msg[0]
		if version != "1.0" {
			panic("AUTH: version != 1.0")
		}

		request_id := msg[1]
		domain := msg[2]
		address := msg[3]
		identity := msg[4]
		mechanism := msg[5]

		username := ""
		password := ""
		client_key := ""
		if mechanism == "PLAIN" {
			username = msg[6]
			password = msg[7]
		} else if mechanism == "CURVE" {
			s := msg[6]
			if len(s) != 32 {
				panic("AUTH: len(client_key) != 32")
			}
			client_key = Z85encode(s)
		}

		allowed := false
		denied := false

		if len(auth_allow) > 0 {
			if auth_allow[address] {
				allowed = true
				if auth_verbose {
					fmt.Printf("AUTH: PASSED (whitelist) address=%q\n", address)
				}
			} else {
				denied = true
				if auth_verbose {
					fmt.Printf("AUTH: DENIED (not in whitelist) address=%q\n", address)
				}
			}
		} else if len(auth_deny) > 0 {
			if auth_deny[address] {
				denied = true
				if auth_verbose {
					fmt.Printf("AUTH: DENIED (blacklist) address=%q\n", address)
				}
			} else {
				allowed = true
				if auth_verbose {
					fmt.Printf("AUTH: PASSED (not in blacklist) address=%q\n", address)
				}
			}
		}

		// Mechanism-specific checks
		if !denied {
			if mechanism == "NULL" && !allowed {
				// For NULL, we allow if the address wasn't blacklisted
				if auth_verbose {
					fmt.Printf("AUTH: ALLOWED (NULL)\n")
				}
				allowed = true
			} else if mechanism == "PLAIN" {
				// For PLAIN, even a whitelisted address must authenticate
				allowed = authenticate_plain(domain, username, password)
			} else if mechanism == "CURVE" {
				// For CURVE, even a whitelisted address must authenticate
				allowed = authenticate_curve(domain, client_key)
			}
		}
		if allowed {
			user_id, metadata := auth_meta_handler(version, request_id, domain, address, identity, mechanism)
			auth_handler.SendMessage(version, request_id, "200", "OK", user_id, metadata)
		} else {
			auth_handler.SendMessage(version, request_id, "400", "NO ACCESS", "", "")
		}
	}

	auth_handler.Close()
}

func authenticate_plain(domain, username, password string) bool {
	for _, dom := range []string{domain, "*"} {
		if m, ok := auth_users[dom]; ok {
			if m[username] == password {
				if auth_verbose {
					fmt.Printf("AUTH: ALLOWED (PLAIN) domain=%q username=%q password=%q\n", dom, username, password)
				}
				return true
			}
		}
	}
	if auth_verbose {
		fmt.Printf("AUTH: DENIED (PLAIN) domain=%q username=%q password=%q\n", domain, username, password)
	}
	return false
}

func authenticate_curve(domain, client_key string) bool {
	for _, dom := range []string{domain, "*"} {
		if m, ok := auth_pubkeys[dom]; ok {
			if m[CURVE_ALLOW_ANY] {
				if auth_verbose {
					fmt.Printf("AUTH: ALLOWED (CURVE any client) domain=%q\n", dom)
				}
				return true
			}
			if m[client_key] {
				if auth_verbose {
					fmt.Printf("AUTH: ALLOWED (CURVE) domain=%q client_key=%q\n", dom, client_key)
				}
				return true
			}
		}
	}
	if auth_verbose {
		fmt.Printf("AUTH: DENIED (CURVE) domain=%q client_key=%q\n", domain, client_key)
	}
	return false
}

// Start authentication.
//
// Note that until you add policies, all incoming NULL connections are allowed
// (classic ZeroMQ behaviour), and all PLAIN and CURVE connections are denied.
func AuthStart() (err error) {
	if auth_init {
		if auth_verbose {
			fmt.Println("AUTH: Already running")
		}
		return errors.New("Auth is already running")
	}

	auth_handler, err = NewSocket(REP)
	if err != nil {
		return
	}
	err = auth_handler.Bind("inproc://zeromq.zap.01")
	if err != nil {
		auth_handler.Close()
		return
	}

	auth_quit, err = NewSocket(REQ)
	if err != nil {
		auth_handler.Close()
		return
	}
	err = auth_quit.Connect("inproc://zeromq.zap.01")
	if err != nil {
		auth_handler.Close()
		auth_quit.Close()
		return
	}

	go auth_do_handler()

	if auth_verbose {
		fmt.Println("AUTH: Starting")
	}

	auth_init = true

	return
}

// Stop authentication.
func AuthStop() {
	if !auth_init {
		if auth_verbose {
			fmt.Println("AUTH: Not running, can't stop")
		}
		return
	}
	if auth_verbose {
		fmt.Println("AUTH: Stopping")
	}
	auth_quit.SendMessage("QUIT")
	auth_quit.RecvMessage(0)
	auth_quit.Close()
	if auth_verbose {
		fmt.Println("AUTH: Stopped")
	}

	auth_init = false

}

// Allow (whitelist) some IP addresses.
//
// For NULL, all clients from these addresses will be accepted.
//
// For PLAIN and CURVE, they will be allowed to continue with authentication.
//
// You can call this method multiple times to whitelist multiple IP addresses.
//
// If you whitelist a single address, any non-whitelisted addresses are treated as blacklisted.
func AuthAllow(addresses ...string) {
	for _, address := range addresses {
		auth_allow[address] = true
	}
}

// Deny (blacklist) some IP addresses.
//
// For all security mechanisms, this rejects the connection without any further authentication.
//
// Use either a whitelist, or a blacklist, not both. If you define both a whitelist
// and a blacklist, only the whitelist takes effect.
func AuthDeny(addresses ...string) {
	for _, address := range addresses {
		auth_deny[address] = true
	}
}

// Add a user for PLAIN authentication for a given domain.
//
// Set `domain` to "*" to apply to all domains.
func AuthPlainAdd(domain, username, password string) {
	if _, ok := auth_users[domain]; !ok {
		auth_users[domain] = make(map[string]string)
	}
	auth_users[domain][username] = password
}

// Remove users from PLAIN authentication for a given domain.
func AuthPlainRemove(domain string, usernames ...string) {
	if u, ok := auth_users[domain]; ok {
		for _, username := range usernames {
			delete(u, username)
		}
	}
}

// Remove all users from PLAIN authentication for a given domain.
func AuthPlainRemoveAll(domain string) {
	delete(auth_users, domain)
}

// Add public user keys for CURVE authentication for a given domain.
//
// To cover all domains, use "*".
//
// Public keys are in Z85 printable text format.
//
// To allow all client keys without checking, specify CURVE_ALLOW_ANY for the key.
func AuthCurveAdd(domain string, pubkeys ...string) {
	if _, ok := auth_pubkeys[domain]; !ok {
		auth_pubkeys[domain] = make(map[string]bool)
	}
	for _, key := range pubkeys {
		auth_pubkeys[domain][key] = true
	}
}

// Remove user keys from CURVE authentication for a given domain.
func AuthCurveRemove(domain string, pubkeys ...string) {
	if p, ok := auth_pubkeys[domain]; ok {
		for _, pubkey := range pubkeys {
			delete(p, pubkey)
		}
	}
}

// Remove all user keys from CURVE authentication for a given domain.
func AuthCurveRemoveAll(domain string) {
	delete(auth_pubkeys, domain)
}

// Enable verbose tracing of commands and activity.
func AuthSetVerbose(verbose bool) {
	auth_verbose = verbose
}

// SUBJECT TO CHANGE
func AuthSetMetaHandler(handler func(version, request_id, domain, address, identity, mechanism string) (user_id string, metadata []byte)) {
	auth_meta_handler = handler
}

//. Additional functions for configuring server or client socket with a single command

// Set NULL server role.
func (server *Socket) ServerAuthNull(domain string) error {
	err := server.SetPlainServer(0)
	if err == nil {
		err = server.SetZapDomain(domain)
	}
	return err
}

// Set PLAIN server role.
func (server *Socket) ServerAuthPlain(domain string) error {
	err := server.SetPlainServer(1)
	if err == nil {
		err = server.SetZapDomain(domain)
	}
	return err
}

// Set CURVE server role.
func (server *Socket) ServerAuthCurve(domain, secret_key string) error {
	err := server.SetCurveServer(1)
	if err == nil {
		err = server.SetCurveSecretkey(secret_key)
	}
	if err == nil {
		err = server.SetZapDomain(domain)
	}
	return err
}

// Set PLAIN client role.
func (client *Socket) ClientAuthPlain(username, password string) error {
	err := client.SetPlainUsername(username)
	if err == nil {
		err = client.SetPlainPassword(password)
	}
	return err
}

// Set CURVE client role.
func (client *Socket) ClientAuthCurve(server_public_key, client_public_key, client_secret_key string) error {
	err := client.SetCurveServerkey(server_public_key)
	if err == nil {
		err = client.SetCurvePublickey(client_public_key)
	}
	if err == nil {
		client.SetCurveSecretkey(client_secret_key)
	}
	return err
}

// SUBJECT TO CHANGE
func AuthMetaBlob(name, value string) ([]byte, error) {
	l1 := len(name)
	if len(name) > 255 {
		return []byte{}, errors.New("Name too long")
	}
	l2 := len(value)
	b := make([]byte, l1+l2+5)
	b[0] = byte(l1)
	b[l1+1] = byte(l2 >> 24 & 255)
	b[l1+2] = byte(l2 >> 16 & 255)
	b[l1+3] = byte(l2 >> 8 & 255)
	b[l1+4] = byte(l2 & 255)
	copy(b[1:], []byte(name))
	copy(b[5+l1:], []byte(value))
	return b, nil
}

// SUBJECT TO CHANGE
func AuthMetaBlobs(pair ...[2]string) ([]byte, error) {
	size := 0
	for _, p := range pair {
		l := len(p[0])
		if l > 255 {
			return []byte{}, errors.New("Name too long")
		}
		size += 5 + l + len(p[1])
	}
	b := make([]byte, 0, size)
	for _, p := range pair {
		b1, _ := AuthMetaBlob(p[0], p[1])
		b = append(b, b1...)
	}
	return b, nil
}
