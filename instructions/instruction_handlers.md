
## handler factor

a handler factory has a function to register handlers

each handler has a name (what the actor is called)


each handler represents an actor with its actions
to understand more about heroscript which is the way how we can actor and actions see @instructions/knowledge/1_heroscript.md

and each handler has as function to translate heroscript  to the implementation

the handler calls the required implementation (can be in one or more packages)

the handler has a play method which uses @pkg/heroscript/playbook to process heroscript and call the required implementation

create a folder in @pkg/heroscript/handlers  which will have all the knowledge how to go from heroscript to implementation


## telnet server

we need a generic telnet server which takes a handler factory as input

the telnet server is very basic, it get's messages
each message is a heroscript

when an empty line is sent that means its the end of a heroscript message

the telnet server needs to be authenticated using a special heroscript message

!!auth secret:'secret123'

as long as that authentication has not been done it will not process any heroscript

the processing of heroscript happens by means of calling the handler factory

there can be more than one secret on the telnet server
