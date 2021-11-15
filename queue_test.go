package bqueue

import (
	"context"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testJob struct {
	processed bool
}

func (j *testJob) Process() {
	j.processed = true
}

func discardLogger() Option {
	return Log(log.New(ioutil.Discard, "", 0))
}

func testQueue(t *testing.T, expectErr bool, options ...Option) {
	t.Helper()

	q, err := New(options...)
	if expectErr {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)

	jobs := make([]*testJob, 5)
	for i := range jobs {
		j := &testJob{}
		jobs[i] = j
		q.Queue(j)
	}

	err = q.Stop(context.Background())
	require.NoError(t, err)

	for _, j := range jobs {
		require.True(t, j.processed)
	}
}

func TestOptions(t *testing.T) {
	cases := []struct {
		name    string
		options []Option
		err     bool
	}{
		{
			name:    "logger",
			options: []Option{discardLogger()},
		},
		{
			name:    "valid-static-workers",
			options: []Option{Static(), Workers(2), discardLogger()},
		},
		{
			name:    "negative-static-workers",
			options: []Option{Static(), Workers(-2), discardLogger()},
			err:     true,
		},
		{
			name:    "zero-static-workers",
			options: []Option{Static(), Workers(0), discardLogger()},
			err:     true,
		},
		{
			name:    "unlimited-static-workers",
			options: []Option{Static(), Workers(UnlimitedWorkers), discardLogger()},
			err:     true,
		},
		{
			name:    "valid-dynamic-workers",
			options: []Option{Workers(2), discardLogger()},
		},
		{
			name:    "negative-dynamic-workers",
			options: []Option{Workers(-2), discardLogger()},
			err:     true,
		},
		{
			name:    "zero-dynamic-workers",
			options: []Option{Workers(0), discardLogger()},
			err:     true,
		},
		{
			name:    "unlimited-dynamic-workers",
			options: []Option{Workers(UnlimitedWorkers), discardLogger()},
		},
		{
			name:    "valid-limit",
			options: []Option{Limit(2), discardLogger()},
		},
		{
			name:    "invalid-limit",
			options: []Option{Limit(-2), discardLogger()},
			err:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testQueue(t, tc.err, tc.options...)
		})
	}
}

type testJobBlocking struct {
}

func (j *testJobBlocking) Process() {
	<-time.After(time.Second)
}

func TestStop(t *testing.T) {
	q, err := New(discardLogger())
	require.NoError(t, err)
	q.Queue(&testJobBlocking{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = q.Stop(ctx)
	require.Error(t, err)
}

func BenchmarkQueue(b *testing.B) {
	q, err := New(Workers(10), Static(), discardLogger())
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		q.Queue(&testJob{})
	}

	err = q.Stop(context.Background())
	require.NoError(b, err)
}
