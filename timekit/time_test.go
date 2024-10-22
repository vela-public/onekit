package timekit

import (
	"fmt"
	"testing"
	"time"
)

func TestDay(t *testing.T) {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1) // 添加0年0月1日，即明天
	fmt.Println("Today:", now)
	fmt.Println("Tomorrow:", tomorrow)
}
