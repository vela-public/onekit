---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by vela.
--- DateTime: 2024/12/17 15:32
---

local hello  = wasm.load("hello.wasm")
local number = hello.add(1,1)
print(number)
