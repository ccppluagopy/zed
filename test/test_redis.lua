local k = KEYS[1]
local k2 = KEYS[2]
local v = ARGV[1]

redis.call("set", k, v)

local ret = redis.call("get", k)

return ret
