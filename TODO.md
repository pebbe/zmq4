## TODO for ZMQ 4.2

see: https://github.com/zeromq/libzmq/releases/tag/v4.2.0
 
### New DRAFT APIs early-release mechanism.

New APIs will be introduced early in public releases, and until they are
stabilized and guaranteed not to change anymore they will be unavailable
unless the new build flag `--enable-drafts` is used. This will allow
developers and early adopters to test new APIs before they are
finalized.

NOTE: as the name implies, NO GUARANTEE is made on the stability of
these APIs. They might change or disappear entirely. Distributions are
recommended NOT to build with them.

New socket types have been introduced in DRAFT state:

    ZMQ_SERVER
	ZMQ_CLIENT
	ZMQ_RADIO
	ZMQ_DISH
	ZMQ_GATHER
	ZMQ_SCATTER
    ZMQ_DGRAM

All these sockets are THREAD SAFE, unlike the existing socket types.
They do NOT support multipart messages (`ZMQ_SNDMORE/ZMQ_RCVMORE`).

`ZMQ_RADIO`, `ZMQ_DISH` and `ZMQ_DGRAM` also support UDP as transport,
both unicast and multicast. See doc/zmq_udp.txt for more details.

New methods to support the new socket types functionality:

    zmq_join
	zmq_leave
	zmq_msg_set_routing_id
	zmq_msg_routing_id
    zmq_msg_set_group
	zmq_msg_group

See doc/zmq_socket.txt for more details.

New poller mechanism and APIs have been introduced in DRAFT state:

    zmq_poller_new
	zmq_poller_destroy
	zmq_poller_add
	zmq_poller_modify
    zmq_poller_remove
	zmq_poller_wait
	zmq_poller_wait_all
	zmq_poller_add_fd
    zmq_poller_modify_fd
	zmq_poller_remove_fd

and a new supporting struct typedef

    zmq_poller_event_t

They support existing socket type, new thread-safe socket types and file
descriptors (cross-platform).

Documentation will be made available in the future before these APIs are
declared stable.

New cross-platform timers helper functions have been introduced in DRAFT
state:

    zmq_timers_new
	zmq_timers_destroy
	zmq_timers_add
	zmq_timers_cancel
    zmq_timers_set_interval
	zmq_timers_reset
	zmq_timers_timeout
    zmq_timers_execute

and a new supporting callback typedef:

    zmq_timer_fn

