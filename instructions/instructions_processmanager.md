

create a process manager

which keeps separate process under control, measures the used cpu, memory

add possibilities to list, create, delete

we can talk to the process manager over local unix domain socket

where we have a super simple protocol we can even test over telnet

## start

```bash
!!process.start name:'processname' command:'command' 

# or
!!process.start name:'processname' 
    command:'command' 

or
!!process.start name:'processname' 
    command:'
        
        ' 

