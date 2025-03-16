in @pkg/vlang

create a vlangprocessor struct which will have some functions

first function  is get_spect(path)

which walks over the path -recursive and finds all .v files
then it will process each of these files

in each file we will look for public structs and public methods on those structs

then return a script which only has

the Struct...
and then the methods on the structs

BUT NO CODE INSIDE THE METHODS

basically the returned codeis just Structs and Methods without the code 

documentation is maintained

test on /Users/despiegk/code/github/freeflowuniverse/herolib/lib/circles/core