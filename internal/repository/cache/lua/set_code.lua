local key=KEYS[1]
local cntKey = key..":cnt"
local value=ARGV[1]

local ttl=tonumber(redis.call("ttl",key))
if ttl==-1 then
    -- 查到key但是没有过期时间
    return -2
elseif ttl==-2 or ttl<540 then
    redis.call("set",key,value)
    redis.call("expire",key,600)
    redis.call("set",cntKey,3)
    redis.call("expire",cntKey,600)
    return 0
else
    -- 发送过于频繁
    return -1
end