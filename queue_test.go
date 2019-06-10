package bqueue

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type aJob struct {
	Name string
}

func (a aJob) Process() error {
	if a.Name != "" {
		log.Printf("test job: %s\n", a.Name)
		return nil
	}

	return errors.New("could not print name")
}

func Test_Collect_and_process_job(t *testing.T) {
	q := New(1)
	q.Start()
	j := aJob{"Boo Foo"}

	output := captureStout(func() {
		q.CollectJob(j)
	})

	assert.Contains(t, output, "test job: Boo Foo", "The two words should be the same.")

}

func Test_WG(t *testing.T) {
	q := New(1)
	q.Start()

	//var output string
	var output string
	output += captureStout(func() {
		for i := 0; i <= 200; i++ {
			j := aJob{strconv.Itoa(i)}
			q.CollectJob(j)

		}
		WG.Wait()
	})

	// look for all instances of "test job: N"
	// not very efficient but minimizes false positive
	for i := 0; i <= 200; i++ {
		if !strings.Contains(output, fmt.Sprintf("test job: %d", i)) {
			t.FailNow()
		}
	}
}

func captureStout(f func()) string {
	var buf bytes.Buffer
	output := ""
	log.SetOutput(&buf)
	start := time.Now()
	f()
	for {
		elapse := time.Since(start).Seconds()
		if elapse > float64(2) {
			break
		}
		output += buf.String()
		if output != "" {
			break
		}
	}

	log.SetOutput(os.Stderr)
	return output
}
