there is a module called doctree

the metadata info of doctree is stored in a redis server

there is the concept of collection

a collection has markdown pages
each page has a unique name in the collection

a collection has files which can be a sstd file or image
each file has a unique name

the names are lower_cased and all non ascci chars are removed, use the namefix function as used in internal/tools/name_fix_test.go 

its a struct called DocTree
which has as argument a path and a name which is also namefixed

the init walks over the path and finds all files and .md files 

we remember the relative position of each file and markdown page in a hset

hset is:

- key: collections:$name
- hkey: pagename.md (always namefixed)
- hkey: imagename.png ... or any other extension for files (always namefixed)
- val: the relative position in the doctree location

use redisclient to internal redis to store this

create following methods on doctree

- Scan (scan the collection again, remove hset and repopulate)
- PageGet get page from a name (do namefix inside method) return the markdown
- PageGetHtml same as PageGet but make html
- FileGetUrl the url which can then be used in static webserver for downloading this content
- PageGetPath relative path in the collection
- Info (name & path)

in PageGet implement a simple include function which is done as !!include name:'pagename' this needs to include page as mentioned in this collection
   if !!include name:'othercollection:pagename' then pagename comes from other collection do namefix to find


## Objects

#### DocTree

- has add, get, delete, list functions in relation to underlying Collection

### Collection

- has get/set/delete/list for pages
- has get/set/delete/list for files

namefix used everywhere to make sure

- in get for page we do the include which an get other pages

## special functions

### Include

```
!!include collectionname:'pagename'
!!include collectionname:'pagename.md'
!!include 'pagename'
!!include collectionname:pagename
!!include collectionname:pagename.md

```

the include needs to parse the following

note:

- pages can have .md or not, check if given if not add
- all is namefixed on collection and page level 
- there can be '' around the name but this is optional