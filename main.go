package main
import (
	"time"
	"strconv"
	_ "net/http/pprof"
	"net/http"
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"runtime/debug"
	"io"
)

type Item int64

type Holder struct {
	objects map[string]*Item
}

func (h *Holder) survive(dur time.Duration) {
	time.Sleep(dur)
}

func main() {
	h1 := Holder{
		objects: make(map[string]*Item),
	}
	h2 := Holder{
		objects: make(map[string]*Item),
	}

	fmt.Println("Setup")

	http.Handle("/debug/pprof/myheap", http.HandlerFunc(func(rw http.ResponseWriter, req*http.Request) {
		f, err := ioutil.TempFile("", "dump")
		if err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(rw, err.Error())
			return
		}
		fmt.Printf("Using %s as heap dump", f.Name())
		defer os.Remove(f.Name())
		defer f.Close()
		debug.WriteHeapDump(f.Fd())
		f.Close()
		f2, err := os.Open(f.Name())
		if err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(rw, err.Error())
			return
		}
		io.Copy(rw, f2)
	}))

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	for i := 0; i < 10000000; i++ {
		it := Item(i)
		s := strconv.FormatInt(int64(i), 10)
		h1.objects[s] = &it
		h2.objects[s] = &it
	}
	fmt.Println("Surviving")
	go h1.survive(time.Second)
	h1.survive(time.Hour * 24)
}
