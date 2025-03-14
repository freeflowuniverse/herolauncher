create a pkg/smtp lib based on 
https://github.com/emersion/go-smtp/tree/master

each mail coming in need to be converted to unicode text
and stored as  json with

from
to
subject
message
attachments []Attachment

Attachment = encoded binary 

into the local redis as hset and in a queue called mail:out which has has the unique id of the mssage
hset is mail:out:$unid -> the json

