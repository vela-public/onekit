# Lua虚拟机

```go
    
    // 创建虚拟机 
    co := lua.NewState()
    co.OpenLibs()
	co.dofile("test.lua")
    co.SetGlobal("a", 1)
	
```

## 参考
项目内核参照了以下项目：
- [gopher-lua](https://github.com/yuin/gopher-lua)