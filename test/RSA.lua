local P = 17
local Q = 19
local N = 323 --[[ N = P * Q ]]
local L = 144 --[[ L = LCM((P - 11), (Q - 1)), LCM是最小公倍数 ]]
local E = 5   --[[ encrypt: (x ^ E) % N, E < L, E为质数, E和L最大公约数为1 ]]
local D = 29  --[[ decrypt: (x ^ D) % N, (E * D) % L = 1 ]]

function encrypt(x)
    local pow = 1
    for i=1, E do
        pow = (pow * (x % N)) % N
    end
    return pow
end

function decrypt(x)
    local pow = 1
    for i=1, D do
        pow = (pow * (x % N)) % N
    end
    return pow
end

for i=1, 12 do
    local n1 = encrypt(i)
    local n2 = decrypt(n1)
    print(n1, n2, i, n2 == i)
end

