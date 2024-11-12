package errkit

import "testing"

func TestErrNo(t *testing.T) {
	flag := Undefined
	flag = flag | Forbidden | Conflict | Failed | Succeed
	t.Log(building.Binary())
	t.Log(Succeed.Binary())
	t.Log(flag.Have(Succeed))
}
