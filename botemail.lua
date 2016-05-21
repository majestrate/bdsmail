--
-- botemail example config
--

bind = ":2525"
domain = "i2p.rocks"
maildir = "/tmp/mail"

function whitelist(addr, recip, sender, body)
   return 0
end

function blacklist(addr, recip, sender, body)
   return 0
end

function checkspam(addr, recip, sender, body)
   return 0
end

