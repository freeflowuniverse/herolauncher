

create a process manager

which keeps separate process under control, measures the used cpu, memory

add possibilities to list, create, delete

we can talk to the process manager over local unix domain socket using a telnet session

## authentication

the telnet server says:

 ** Welcome: you are not authenticated, provide secret.

then we pass the secret which was passed when we started the process manager

once authenticated it says

** Welcome: you are authenticated.

now we can send heroscripts to it (see @pkg/playbook for how to parse that)

## actions can be sent over telnet

just send heroscript statements

everytime a new !! goes or # as comment we execute the previous heroscripts

## we make handlers 

using the playbook: @pkg/playbook

this checks which commands are sent and this then calls the corresponding handler and instructs the processmanager

## start

heroscript

```bash
!!process.start name:'processname' command:'command\n which can be multiline' log:true 
    deadline:30 cron:'0 0 * * *'  jobid:'e42'

```


## list


heroscript

```bash
!!process.list format:json

```

lists the processes and returns as json

when telnet protocol needs to return its always as 

**RESULT** e42
... here is the result in chosen format
**ENDRESULT**

if jobid specified on the heroscript action then its shown behind **RESULT** if not then its empty

## delete

```bash
!!process.delete name:'processname'

```

## status

```bash
!!process.status name:'processname' format:json

```

shows mem  usage, cpu usage, status e.g. running ...

## restart, stop, start

do same as status but then for these

## log


```bash
!!process.log name:'processname' format:json limit:100

```

returns the last 100 lines of the log

if not format then just the log itself


