package mandala

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"github.com/tideland/goas/v2/loop"
)

const (
	// The number of event messages buffered by the event
	// channel. This is an heuristic value.
	NumOfBufferedEvents = 10
)

var (
	// If Verbose is true Logf will print on the stdout.
	Verbose bool

	// If Debug is true Debugf will print on the stdout.
	Debug bool

	// Send commands to this channel in order to manage
	// application's resources.
	request chan interface{}

	// This is the internal global event channel. All system
	// events should be sent to this channel.
	event chan interface{}

	// Internal requests to the sound engine are sent to this
	// channel.
	soundCh chan interface{}

	// The current activity pointer is sent to this channel when
	// an onCreate event is triggered by Android.
	activity chan unsafe.Pointer
)

// Fatalf simply calls log.Fatalf
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

// If Verbose is true then Logf will print on stdout.
func Logf(format string, v ...interface{}) {
	if Verbose {
		log.Printf(format, v...)
	}
}

// If Debug is true then Debugf will print on stdout.
func Debugf(format string, v ...interface{}) {
	if Debug {
		log.Printf(format, v...)
	}
}

// Stacktrace returns the stacktrace of the goroutine that calls it.
func Stacktrace() string {
	// Write a stack trace
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, true)

	// Incrementally grow the
	// buffer as the stack trace
	// requires.
	for n > len(buf) {
		buf = make([]byte, len(buf)*2)
		n = runtime.Stack(buf, false)
	}
	return string(buf)
}

// ResourceManager returns a send-only channel to which client-code
// send request for resources. Please refer to resourcemanager.go for a
// complete list of supported requests.
func ResourceManager() chan<- interface{} {
	return request
}

// Events returns a receive-only channel from which client-code
// receive events. Events are sent in the form of anonymous
// interfaces. Please refer to events.go for a complete list of the
// supported events. For a worthwhile reading take a look a this
// document by NVIDIA:
//
// http://developer.download.nvidia.com/assets/mobile/files/AndroidLifecycleAppNote_v100.pdf
//
func Events() <-chan interface{} {
	return event
}

func init() {
	event = make(chan interface{}, NumOfBufferedEvents)
	request = make(chan interface{})
	activity = make(chan unsafe.Pointer, 1)
	soundCh = make(chan interface{})

	loop.GoRecoverable(
		resourceLoopFunc(activity, request),
		func(rs loop.Recoverings) (loop.Recoverings, error) {
			for _, r := range rs {
				Logf("%s", r.Reason)
				Logf("%s", Stacktrace())
			}
			return rs, fmt.Errorf("Unrecoverable loop\n")
		},
	)
}
