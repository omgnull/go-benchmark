package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/oxtoacart/bpool"
	"github.com/valyala/bytebufferpool"
)

type (

	// Program config
	Config struct {
		// number of goroutines
		queue int

		// function to run
		method string

		// filename to write report
		out string

		// Run duration
		duration time.Duration
	}

	// Method type
	Fn func()
)

var (
	wg sync.WaitGroup

	// Number of methods runs
	runs uint64

	// Default config
	config = &Config{
		queue:    runtime.NumCPU(),
		method:   "generic",
		duration: 60 * time.Second,
	}

	// Queue size
	queue = make(chan byte)

	// Allowed methods
	methods = map[string]Fn{
		"generic": GenericBuf,
		"stack":   GenericStackBuf,
		"alloc":   AllocBuf,
		"sync":    SyncBuf,
		"bpool":   BpoolBuf,
		"bbpool":  BBpoolBuf,
	}

	// Strings written to buf
	str = []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
		`Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris
		nisi ut aliquip ex ea commodo consequat.
		Duis aute irure dolor in reprehenderit in voluptate velit esse cillum
		dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident,
		sunt in culpa qui officia deserunt mollit anim id est laborum`,
		"Sed ut perspiciatis",
		"sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt",
		"Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit",
		"laboriosam, nisi ut aliquid ex ea commodi consequatur",
		"Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur",
		"vel illum qui dolorem eum fugiat quo voluptas nulla pariatur",
	}

	// Generic size
	sPool = &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}

	// Default size
	bPool = bpool.NewBufferPool(20)
)

// Simulate const handling
func main() {
	var duration float64

	flag.IntVar(&config.queue, "queue", config.queue, "Number of goroutines; default is NumCPU")
	flag.StringVar(&config.method, "method", config.method, fmt.Sprintf("Function to run; allowed:%v", Methods()))
	flag.StringVar(&config.out, "out", config.out, "Filename to write report; Prints into stdout by default")
	flag.Float64Var(&duration, "duration", float64(config.duration.Seconds()), "Test duration in seconds")
	flag.Parse()

	config.queue = int(math.Max(1, float64(config.queue)))
	config.duration = time.Duration(math.Max(5, duration)) * time.Second

	run(config)
}

func run(c *Config) {
	queue = make(chan byte, config.queue)
	signals := make(chan os.Signal, 1)
	reports := make(chan byte, 1)

	method := GetMethod(config.method)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	start := time.Now()

	if c.out != "" {
		go MakeReport(start, c.out, reports)
	} else {
		go PrintReport(start, reports)
	}

LOOP:
	for {
		select {
		case <-signals:
			// Stop
			reports <- 0
			break LOOP
		case queue <- 0:
			wg.Add(1)
			atomic.AddUint64(&runs, 1)
			go method()
		default:
			if time.Since(start).Seconds() >= config.duration.Seconds() {
				// Stop
				reports <- 0
				break LOOP
			}
		}
	}

	wg.Wait()
}

func GetMethod(name string) Fn {
	fn, ok := methods[name]
	if !ok {
		log.Fatalf("Could not find method [%s]; allowed methods are: %v", name, Methods())
	}
	return fn
}

func Methods() string {
	var out string
	for k := range methods {
		out += fmt.Sprintf(" %q", k)
	}
	return out
}

func WorkWithBuf(b *bytes.Buffer) {
	for _, s := range str {
		b.WriteString(s)
	}
}

func WorkWithByteBuf(b *bytebufferpool.ByteBuffer) {
	for _, s := range str {
		b.WriteString(s)
	}
}

func GenericStackBuf() {
	var buf bytes.Buffer
	buf.Reset()
	for _, s := range str {
		buf.WriteString(s)
	}
	Done()
}

func GenericBuf() {
	var buf bytes.Buffer
	buf.Reset()
	WorkWithBuf(&buf)
	Done()
}

func AllocBuf() {
	buf := new(bytes.Buffer)
	buf.Reset()
	WorkWithBuf(buf)
	Done()
}

func SyncBuf() {
	buf := sPool.Get().(*bytes.Buffer)
	WorkWithBuf(buf)
	buf.Reset()
	sPool.Put(buf)
	Done()
}

func BpoolBuf() {
	buf := bPool.Get()
	WorkWithBuf(buf)
	bPool.Put(buf)
	Done()
}

func BBpoolBuf() {
	buf := bytebufferpool.Get()
	WorkWithByteBuf(buf)
	bytebufferpool.Put(buf)
	Done()
}

func Done() {
	select {
	case <-queue:
		wg.Done()
	default:
		// skip
	}
}

func MakeReport(start time.Time, filepath string, reports chan byte) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	m := new(runtime.MemStats)
	w := bufio.NewWriter(file)

	// Write headers
	file.WriteString("Time;Runs;Alloc;TotalAlloc;Mallocs;Frees;HeapAlloc;HeapSys;HeapIdle;HeapReleased;StackInuse\n")

REPORT_FILE:
	for {
		select {
		case <-reports:
			break REPORT_FILE
		default:
			runtime.ReadMemStats(m)

			// Elapsed time
			w.WriteString(strconv.FormatInt(int64(time.Since(start).Seconds()), 10))
			w.WriteByte(';')

			// Number of runs
			w.WriteString(strconv.FormatUint(runs, 10))
			w.WriteByte(';')

			// ------------- ALLOC
			// Alloc
			w.WriteString(strconv.FormatUint(m.Alloc, 10))
			w.WriteByte(';')

			// TotalAlloc
			w.WriteString(strconv.FormatUint(m.TotalAlloc, 10))
			w.WriteByte(';')

			// Mallocs
			w.WriteString(strconv.FormatUint(m.Mallocs, 10))
			w.WriteByte(';')

			// Frees
			w.WriteString(strconv.FormatUint(m.Frees, 10))
			w.WriteByte(';')
			// ------------- ALLOC

			// ------------- HEAP
			// HeapAlloc
			w.WriteString(strconv.FormatUint(m.HeapAlloc, 10))
			w.WriteByte(';')

			// HeapSys
			w.WriteString(strconv.FormatUint(m.HeapSys, 10))
			w.WriteByte(';')

			// HeapIdle
			w.WriteString(strconv.FormatUint(m.HeapIdle, 10))
			w.WriteByte(';')

			// HeapReleased
			w.WriteString(strconv.FormatUint(m.HeapReleased, 10))
			w.WriteByte(';')
			// ------------- HEAP

			// ------------- STACK
			// Alloc
			w.WriteString(strconv.FormatUint(m.StackInuse, 10))
			w.WriteByte('\n')
			// ------------- ALLOC

			w.Flush()
			time.Sleep(time.Second)
		}
	}
}

func PrintReport(start time.Time, reports chan byte) {
	m := new(runtime.MemStats)

REPORT:
	for {
		select {
		case <-reports:
			break REPORT
		default:
			runtime.ReadMemStats(m)

			// Elapsed time
			fmt.Printf("%v;", int64(time.Since(start).Seconds()))

			// Number of runs
			fmt.Printf("%v;", runs)

			// ------------- ALLOC
			// Alloc
			fmt.Printf("%v;", m.Alloc)

			// TotalAlloc
			fmt.Printf("%v;", m.TotalAlloc)

			// Mallocs
			fmt.Printf("%v;", m.Mallocs)

			// Frees
			fmt.Printf("%v;", m.Frees)
			// ------------- ALLOC

			// ------------- HEAP
			// HeapAlloc
			fmt.Printf("%v;", m.HeapAlloc)

			// HeapSys
			fmt.Printf("%v;", m.HeapSys)

			// HeapIdle
			fmt.Printf("%v;", m.HeapIdle)

			// HeapReleased
			fmt.Printf("%v;", m.HeapReleased)
			// ------------- HEAP

			// ------------- STACK
			// StackInuse
			fmt.Printf("%v;", m.StackInuse)
			// ------------- ALLOC

			fmt.Print("\n")
			time.Sleep(time.Second)
		}
	}
}
