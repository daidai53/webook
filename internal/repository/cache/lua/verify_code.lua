local key=KEYS[1]
local cntKey = key..":cnt"
local expectCode=ARGV[1]

local cnt=tonumber(redis.call("get",cntKey))
local code =redis.call("get",key)

if cnt==nil or cnt<=0 then
    -- 验证次数耗尽
    return -1
end

if code==expectCode then
    redis.call("set",cntKey,0)
    return 0
else
    redis.call("decr",cntKey)
    return -2
end