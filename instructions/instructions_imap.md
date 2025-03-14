create a pkg/imapserver lib which uses:

https://github.com/foxcpp/go-imap

the mails are in redis

the model for mail is in @pkg/mail/model.go

## the mails are in redis based on following code, learn from it

cmd/redis_mail_feeder/main.go

the redis keys are

- mail:in:$account:$folder:$uid

the json is the mail model

see @instructions_imap_feeder.md for details

## imap server is using the redis as backedn

- based on what the feeder put in

there is no no login/passwd, anything is fine, any authentication is fine,
ignore if user specifies it, try to support any login/passwd/authentication method just accept everything


