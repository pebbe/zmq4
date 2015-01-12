package zmq4

/*
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <netdb.h>

char *my_get_peer_addr(int srcFd) {
    struct sockaddr_storage
        ss;
    socklen_t
        addrlen;
    int
        rc;
    char
        host [NI_MAXHOST];

    // get the remote endpoint
    addrlen = sizeof ss;
    rc = getpeername (srcFd, (struct sockaddr*) &ss, &addrlen);
    if (rc != 0)
        return NULL;

    rc = getnameinfo ((struct sockaddr*) &ss, addrlen, host, sizeof host, NULL, 0, NI_NUMERICHOST);
    if (rc != 0)
        return NULL;

    return strdup(host);
}
*/
import "C"

import (
	"unsafe"
)

// Given a socket fd, return the peer address.
func GetPeerAddr(fd int) (string, error) {
	s, err := C.my_get_peer_addr(C.int(fd))
	if s == nil {
		return "", errget(err)
	}
	defer C.free(unsafe.Pointer(s))
	return C.GoString(s), nil
}
