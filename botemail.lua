--
-- botemail example config
--

bind = ":0"
domain = "i2p.rocks"
maildir = "/tmp/mail"

function whitelist(recip, sender, body)
   return 1
end

function blacklist(recip, sender, body)
   return 0
end

function checkspam(recip, sender, body)
   return 0
end

