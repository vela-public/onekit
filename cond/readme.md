# condition

> 表达式快速匹配模块 解决 过于冗长的条件判断 简化 开发逻辑

## vela.cnd(string , string , string)

- 编译一个条件表示的功能模块
- cond 实现了 NewIndex的方法 支持直接赋值修改
- ok 匹配成功后触发相关逻辑 类型:pipe 参数:match传入的参数
- no 没又匹配后触发相关逻辑 类型:pipe 参数:match传入的参数
- match(object) 匹配

```lua
    local cnd = vela.cnd("name eq zhang1,zhang2,zhang3")
    
    cnd.ok(function(v) print(v.name , v.pass) end)
    cnd.no(function(v) print(v.name , v.pass) end)
    
    local object = {
        name = "wangwu",
        pass = "123456",
    }
    cnd.match(object)
```

## vela.cndf(prefix , value , value..)

> 快速生成条件

```lua
    local c1 = vela.cndf("key = " , 10 , 20)
    local c2 = vela.cndf("ext = " , ".json" , ".txt")

    print(c1) -- key = 10,20
    print(c2) -- key = .json,.txt
```

## 条件格式

> 格式内容

key,key,key &emsp;键值采用逗号分割 代表多个 <br />
val,val,val &emsp;匹配那些值也可以采用逗号 代表多个

**中间 为 表达是 缩写 包含 eq,re,cn,in,lt,le,ge,gt,->**

- eq &emsp;等于规则(aa eq aa,bb,cc) 符号: =
- re &emsp;匹配规则(abc re a*,\*b\*,*c) 只有星号和点 不支持正则表达式
- cn &emsp;包含内容(abc cn *c)
- in &emsp;等于包含(1 in 1,3,4)
- lt &emsp;小于
- le &emsp;小于等于
- ge &emsp;大于等于
- gt &emsp;大于
- -> &emsp;call

```bash
    # 单个
    key eq value 
    key eq value,value,value
    
    # 多个
    key1,key2 eq value1,value2
    
    # 取反
    key !eq value,value2,value3
    key !cn *v,*v2
    
    key -> /risk/ip?have&kind=tor 
```

## string

> string 默认自带内置key说明

```lua
    --匹配对象为string 12-345-678.raw 
    cond("*     = 123") --完全匹配
    cond("ext   = .jar,.aa") --是否有后缀
    cond("ipv4  = true") --是否为IPv4
    cond("ipv6  = true") --是否为IPv4
    cond("ip    = true") --是否为IPv4
    cond("[1:3] = 2-3")
    cond("[1]  =  #") --判断是否为#

```