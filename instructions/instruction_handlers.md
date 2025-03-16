
## handler factor

a handler factory has a function to register handlers


each handler represents an actor with its actions
to understand more about heroscript which is the way how we can actor and actions see @instructions/knowledge/1_heroscript.md

and each handler has as function to translate heroscript  to the implementation

the handler calls the required implementation (can be in one or more packages)

the handler has a play method which uses @pkg/heroscript/playbook to process heroscript and call the required implementation

## telnet server

we need a generic telnet server

which we can add handlers to by means of a HandlerFactory

a HandlerFactory has multiple handlers

each handler is a Struct with methods 

each method accept heroscript as input (string)

heroscript see @instructions/knowledge/1_heroscript.md 

inside this method we then process the heroscript using  @pkg/heroscript

create a pkg called handlerfactory

in cmd in this package create an example handler which gets attached to the factory

each handler processes the actions for an actor

so when defining a handler we specify the name of the actor

each method on the handler is the action of the actor

