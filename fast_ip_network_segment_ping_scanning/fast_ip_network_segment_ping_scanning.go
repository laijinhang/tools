package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

var (
	icmp  ICMP
	laddr = net.IPAddr{IP: net.ParseIP("ip")}
	num     int = 4
	timeout int64 = 1500
	size    int = 32
	stop    bool = false
)

func main() {
	bIp := []int{192, 168, 1, 1}	// 开始ip
	eIp := []int{192, 168, 1, 255}	// 结束ip

	wg := sync.WaitGroup{}
	number := make(chan bool, 5000)
	for i1 := bIp[0];i1 <= eIp[0];i1++ {
		for i2 := bIp[1];i2 <= eIp[1];i2++ {
			for i3 := bIp[2];i3 <= eIp[2];i3++ {
				for i4 := bIp[3];i4 <= eIp[3];i4++ {
					number <- true
					wg.Add(1)
					go func(i1, i2, i3, i4 int) {
						defer func() {
							<-number
							wg.Done()
						}()
						desIp := fmt.Sprintf("%d.%d.%d.%d", i1, i2, i3, i4)
						if ping(desIp) {
							fmt.Println(desIp)
						}
					}(i1, i2, i3, i4)
				}
			}
		}
	}
	wg.Wait()
}

func ping(desIp string) bool {
	conn, err := net.DialTimeout("ip4:icmp", desIp, time.Duration(timeout) * time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	//icmp头部填充
	icmp.Type = 8
	icmp.Code = 0
	icmp.Checksum = 0
	icmp.Identifier = 1
	icmp.SequenceNum = 1

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp) // 以大端模式写入
	data := make([]byte, size)                    //
	buffer.Write(data)
	data = buffer.Bytes()

	var SuccessTimes int	// 成功次数
	var FailTimes int		// 失败次数
	var minTime	int = int(math.MaxInt32)
	var maxTime int
	var totalTime int
	for i := 0;i < num;i++ {
		icmp.SequenceNum = uint16(1)
		// 检验和设为0
		data[2] = byte(0)
		data[3] = byte(0)

		data[6] = byte(icmp.SequenceNum >> 8)
		data[7] = byte(icmp.SequenceNum)
		icmp.Checksum = CheckSum(data)
		data[2] = byte(icmp.Checksum >> 8)
		data[3] = byte(icmp.Checksum)

		// 开始时间
		t1 := time.Now()
		conn.SetDeadline(t1.Add(time.Duration(time.Duration(timeout) * time.Millisecond)))
		_, err := conn.Write(data)
		if err != nil {
			log.Fatal(err)
		}
		buf := make([]byte, 65535)
		_, err = conn.Read(buf)
		if err != nil {
			FailTimes++
			continue
		}
		et := int(time.Since(t1) / 1000000)
		if minTime > et {
			minTime = et
		}
		if maxTime <et {
			maxTime = et
		}
		totalTime += et
		SuccessTimes++
		break
	}
	if SuccessTimes != 0 {
		return true
	}
	return false
}

func CheckSum(data []byte) uint16 {
	var sum uint32
	var length = len(data)
	var index int
	for length > 1 { // 溢出部分直接去除
		sum += uint32(data[index]) << 8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length == 1 {
		sum += uint32(data[index])
	}
	// CheckSum的值是16位，计算是将高16位加低16位，得到的结果进行重复以该方式进行计算，直到高16位为0
	/*
		sum的最大情况是：ffffffff
		第一次高16位+低16位：ffff + ffff = 1fffe
		第二次高16位+低16位：0001 + fffe = ffff
		即推出一个结论，只要第一次高16位+低16位的结果，再进行之前的计算结果用到高16位+低16位，即可处理溢出情况
	*/
	sum = uint32(sum >> 16) + uint32(sum)
	sum = uint32(sum >> 16) + uint32(sum)
	return uint16(^sum)
}
