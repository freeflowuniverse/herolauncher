in @pkg/system/stats 
create a factory which is called StatsManager

then each method in the different files is a method on that StatsManager

then on StatsManager make a connection to a redis server and have this connection as property

then have a dict on the StatsManager called Expiration which in seconds defines per type of info how long we will cache

then on each method cache the info in redis on a well chosen key, if someone asks the info and its out of expiration then we send message to the goroutine which fetches the info (see further) for the next request, in other words we will stil give the info we had in cache but next request would then have the new info if it got fetched in time

make a goroutine which is doing the updates, only when info is asked for it will request it,
have an internal queue which tells this go-routine which info to ask for, NO info can be asked in parallel its one after the other

when an external consumer asks for the info it always comes from the cache

when the system starts up, the goroutine will do a first fetch, so in background initial info is loaded

when a request is asked from external consumer and info is not there yet, the method wll keep on polling redis this info is there (block wait), this redis will only be filled in once the goroutine fetches it

there is a generic timeout of 1 min

put a debug flag (bool) on the StatsManager if that one is set then the request for stats is always direct, not waiting for cache and no go-routine used