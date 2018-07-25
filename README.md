BQUEUE
----
**A buffered async queue**
###### Based on Marcio Castilho [article](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/) “Handling 1 Million Requests per Minute with Go”

Why
---
We needed a simple and quick queue system. to handle request rapidely without impacting performance.
Each of our job could take up to 20s to execute.

How it works
----
by using the awesomeness of channels

>Channels are the pipes that connect concurrent goroutines. You can send values into channels from one goroutine and receive those values into another goroutine

With channels we are able to flow the data between states without any performance impact

```
  ┌─────────────────────────────────┐                             ┌───────────────────────────────┐
  │Queue                            │                             │Worker                         │
  │---                              │                             │---                            │
  │    maxWorker   int              │                             │    ID          int            │
  │    JobRequests chan chan Job    │                             │    JobHunting  chan Job       │
  │    JobReceived chan Job         │                             │    jobRequests chan chan Job  │
  │                                 │                             │    QuitChan    chan bool      │
  │                                 │                             │                               │
  │M: Start()                       │                             │                               │
  │-                                │                             │                               │
  │Also initialize the workers      │     ┌────────────────┐      │                               │
  │and channels                     │     │ share the same │      │M: Start() |  goroutine        │
  │                                 │     │    instance    │      │-                              │
  │                                 │     ├────────────────┤  ┌──▶│Receive a job from JobRequests │
  │M: CollectJob(j Job)             │     │JobRequests     │  │   │Calls do passing the job       │
  │-                                │┌───▶│---             │──┘   │                               │
  │adds the job to the JobReceived  ││    │(chan chan Job) │      │                               │
  │                                 ││    └────────────────┘      │                               │
  │                                 ││                            │M: do(j Job)                   │
  │M: dispatch() | goroutine      ──┼┘                            │-                              │
  │-                                │                             │process the job                │
  │Looks for job in JobReceived     │                             │                               │
  │Then add it to the JobRequests   │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  │                                 │                             │                               │
  └─────────────────────────────────┘                             └───────────────────────────────┘
```
