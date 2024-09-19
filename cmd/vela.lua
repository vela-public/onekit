print("hello")

local tab = luakit.map{
	name = "vela",
	age = 10,
	10,
	2,
}

tab.love = "security"

print(luakit.fmt("%v %v %s" , "hello" , 5 , tab.love))


local arr = luakit.slice("123" , 345)

print(arr[1])


decode{
	id =  123,
	name = "vela.name",
	addr = {"123" , "123" , "123" , "456" , "345"}
}