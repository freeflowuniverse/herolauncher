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

