local key = KEYS[1]
local member = KEYS[2]
local delta = tonumber(ARGV[1])

redis.call("ZINCRBY",key,delta,member)
