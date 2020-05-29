#if ZMQ_VERSION_MAJOR != 4

#error "You need ZeroMQ version 4 to build this"

#endif

#if ZMQ_VERSION_MINOR < 1

#define ZMQ_CONNECT_RID -1
#define ZMQ_GSSAPI -1
#define ZMQ_GSSAPI_PLAINTEXT -1
#define ZMQ_GSSAPI_PRINCIPAL -1
#define ZMQ_GSSAPI_SERVER -1
#define ZMQ_GSSAPI_SERVICE_PRINCIPAL -1
#define ZMQ_HANDSHAKE_IVL -1
#define ZMQ_IPC_FILTER_GID -1
#define ZMQ_IPC_FILTER_PID -1
#define ZMQ_IPC_FILTER_UID -1
#define ZMQ_ROUTER_HANDOVER -1
#define ZMQ_SOCKS_PROXY -1
#define ZMQ_THREAD_PRIORITY -1
#define ZMQ_THREAD_SCHED_POLICY -1
#define ZMQ_TOS -1
#define ZMQ_XPUB_NODROP -1

#endif

#if ZMQ_VERSION_MINOR < 2

#define ZMQ_MAX_MSGSZ -1

#define ZMQ_BLOCKY -1
#define ZMQ_XPUB_MANUAL -1
#define ZMQ_XPUB_WELCOME_MSG -1
#define ZMQ_STREAM_NOTIFY -1
#define ZMQ_INVERT_MATCHING -1
#define ZMQ_HEARTBEAT_IVL -1
#define ZMQ_HEARTBEAT_TTL -1
#define ZMQ_HEARTBEAT_TIMEOUT -1
#define ZMQ_XPUB_VERBOSER -1
#define ZMQ_CONNECT_TIMEOUT -1
#define ZMQ_TCP_MAXRT -1
#define ZMQ_THREAD_SAFE -1
#define ZMQ_MULTICAST_MAXTPDU -1
#define ZMQ_VMCI_BUFFER_SIZE -1
#define ZMQ_VMCI_BUFFER_MIN_SIZE -1
#define ZMQ_VMCI_BUFFER_MAX_SIZE -1
#define ZMQ_VMCI_CONNECT_TIMEOUT -1
#define ZMQ_USE_FD -1

#define ZMQ_GROUP_MAX_LENGTH -1

#define ZMQ_POLLPRI -1

#endif

#if ZMQ_VERSION_MINOR == 2 && ZMQ_VERSION_PATCH < 3

#define ZMQ_EVENT_HANDSHAKE_FAILED_NO_DETAIL -1
#define ZMQ_EVENT_HANDSHAKE_SUCCEEDED -1
#define ZMQ_EVENT_HANDSHAKE_FAILED_PROTOCOL -1
#define ZMQ_EVENT_HANDSHAKE_FAILED_AUTH -1

#endif

#ifndef ZMQ_ROUTING_ID
#define ZMQ_ROUTING_ID ZMQ_IDENTITY
#endif
#ifndef ZMQ_CONNECT_ROUTING_ID
#define ZMQ_CONNECT_ROUTING_ID ZMQ_CONNECT_RID
#endif

int zmq4_bind (void *socket, const char *endpoint);
int zmq4_close (void *socket);
int zmq4_connect (void *socket, const char *endpoint);
int zmq4_ctx_get (void *context, int option_name);
void *zmq4_ctx_new (void);
int zmq4_ctx_set (void *context, int option_name, int option_value);
int zmq4_ctx_term (void *context);
int zmq4_curve_keypair (char *z85_public_key, char *z85_secret_key);
int zmq4_curve_public (char *z85_public_key, char *z85_secret_key);
int zmq4_disconnect (void *socket, const char *endpoint);
int zmq4_getsockopt (void *socket, int option_name, void *option_value, size_t *option_len);
const char *zmq4_msg_gets (zmq_msg_t *message, const char *property);
int zmq4_msg_recv (zmq_msg_t *msg, void *socket, int flags);
int zmq4_poll (zmq_pollitem_t *items, int nitems, long timeout);
int zmq4_proxy (const void *frontend, const void *backend, const void *capture);
int zmq4_proxy_steerable (const void *frontend, const void *backend, const void *capture, const void *control);
int zmq4_send (void *socket, void *buf, size_t len, int flags);
int zmq4_setsockopt (void *socket, int option_name, const void *option_value, size_t option_len);
void *zmq4_socket (void *context, int type);
int zmq4_socket_monitor (void *socket, char *endpoint, int events);
int zmq4_unbind (void *socket, const char *endpoint);

/* DRAFT */

int zmq4_msg_init_size (zmq_msg_t *msg, size_t size);
int zmq4_msg_send (zmq_msg_t *msg, void *socket, int flags);
int zmq4_msg_set_group (zmq_msg_t *msg, const char *group);
int zmq4_msg_set_routing_id (zmq_msg_t *message, uint32_t routing_id);
int zmq4_join (void *s, const char *group);
int zmq4_leave (void *s, const char *group);
