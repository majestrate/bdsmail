--
-- botemail example config
--

bind = ":2525"
domain = "i2p.rocks"
maildir = "/tmp/mail"


--
-- the following functions are required to be defined
--


--
-- return 1 if this mail is whitelisted otherwise return 0
--
function whitelist(addr, recip, sender, body)
   return 0
end

--
-- return 1 if this mail is blacklisted otherwise return 0
--
function blacklist(addr, recip, sender, body)
   return 0
end

--
-- return 1 if this mail is spam otherwise return 0
--
function checkspam(addr, recip, sender, body)
   return 0
end

