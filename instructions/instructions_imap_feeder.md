
## populator of imap

in @/cmd/ 
create a new command called redis_mail_feeder

this feeder creates 100 mails in different folders and stores them in redis as datastor

the mail model is in @pkg/mail/model.go

@uid is epoch in seconds + an incrementing number based on of there was already a mail with the same uid before, so we just increment the number  until we get a unique uid (is string(epoch)+string(incrementing number))

the mails are stored in

and stores mail in mail:in:$account:$folder:$uid

id is the blake192 from the json serialization 

- account is random over pol & jan
- folders chose then random can be upto 3 levels deep

make random emails, 100x in well chosen folder

