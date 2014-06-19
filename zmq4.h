#if ZMQ_VERSION_MINOR > 0
typedef struct {
    uint16_t event;  // id of the event as bitfield
    int32_t  value ; // value is either error code, fd or reconnect interval
} zmq_event_t;
#endif
