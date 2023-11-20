local key=KEYS[1]
local window = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local min=now-window

-- 删除不在时间窗口的记录
redis.call("ZREMRANGEBYSCORE",key,"-inf",min)
local count=redis.call("ZCOUNT",key,min,now)
if count>=threshold then
    return 'true'
else
    redis.call("ZADD",key,now,now)
    redis.call("PEXPIRE",key,window)
    return 'false'
end