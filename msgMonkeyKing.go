package main

import "fmt"
import "net"
import "os"
import "strconv"
import "time"
import "strings"

func LogMsg(v ...interface{}) {
	fmt.Println(v...)
}

func UpStreamReader(conn net.Conn, dataChan chan []byte, controlChan chan string) {

	buffer := make([]byte, 1024)

	for {

		size, error1 := conn.Read(buffer)

		if error1 != nil {

			controlChan <- "close"

			return
		}

		dataChan <- buffer[:size]

	}
}

func DownStreamReader(conn net.Conn, dataChan chan []byte, controlChan chan string) {

	buffer := make([]byte, 1024)

	for {

		size, error1 := conn.Read(buffer)

		if error1 != nil {

			controlChan <- "close"

			return

		}

		dataChan <- buffer[:size]

	}
}

type RelayData struct {
	Buff []byte

	OK bool
}

func CreateBuffByDiscard(totalBuff []byte, discarded int) []byte {

	totalSize := len(totalBuff)

	remainBuff := make([]byte, totalSize-discarded)

	copy(remainBuff, totalBuff[discarded:])

	return remainBuff
}

func SendData(conn net.Conn, buffer []byte) (bool, int) {

	isOK := true

	doneSize, err := conn.Write([]byte(buffer))

	if err != nil {

		isOK = false

	}

	return isOK, doneSize
}

type RelayUpStreamDataCallBack func(relayData *RelayData, conn net.Conn, segSize int, flush bool)

func RelayUpStreamDataByMerge(relayData *RelayData, conn net.Conn, segSize int, flush bool) {
	// this function need to be optimized( to use the ring buffer),
	// currently just make it as simple as possible	
	totalSize := len(relayData.Buff)

	if totalSize == 0 {

		return
	}

	if flush {

		isOK, doneSize := SendData(conn, relayData.Buff)

		//when we do flush, if we can not send all data out,
		//take it as the error case
		if !isOK || doneSize != totalSize {

			relayData.OK = false

			return
		}

		relayData.Buff = CreateBuffByDiscard(relayData.Buff, doneSize)

		return
	}

	if totalSize < segSize {
		// no flush and totalSize is NOT bigger enough, so do nothing
		return
	}

	//when it reach here, means totalSize >= segSize
	isOK, doneSize := SendData(conn, relayData.Buff[:segSize])

	if !isOK {

		relayData.OK = false

		return
	}

	relayData.Buff = CreateBuffByDiscard(relayData.Buff, doneSize)

	return
}

func RelayUpStreamDataBySplit(relayData *RelayData, conn net.Conn, segSize int, flush bool) {

	// this function need to be optimized( to use the ring buffer),
	// currently just make it as simple as possible	

	totalSize := len(relayData.Buff)

	if totalSize == 0 {

		return
	}

	if flush {

		isOK, doneSize := SendData(conn, relayData.Buff)

		//when we do flush, if we can not send all data out,
		//take it as the error case
		if !isOK || doneSize != totalSize {

			relayData.OK = false

			return
		}

		relayData.Buff = CreateBuffByDiscard(relayData.Buff, doneSize)

		return
	}

	if totalSize <= segSize {

		isOK, doneSize := SendData(conn, relayData.Buff)

		if !isOK {

			relayData.OK = false

			return
		}

		relayData.Buff = CreateBuffByDiscard(relayData.Buff, doneSize)

		return
	}

	isOK, doneSize := SendData(conn, relayData.Buff[:segSize])

	if !isOK {

		relayData.OK = false

		return
	}

	relayData.Buff = CreateBuffByDiscard(relayData.Buff, doneSize)

	return
}

func Moneky(conn net.Conn, dstAddr string, dstPort string, hander RelayUpStreamDataCallBack, segSize int, timeout_ms int) {

	addr, err := net.ResolveTCPAddr("tcp", dstAddr+":"+dstPort)

	if err != nil {
		panic(0)
	}

	tconn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(0)
	}

	tconn.SetNoDelay(true)

	upStreamDataChannel := make(chan []byte)
	upStreamControlChannel := make(chan string)

	go UpStreamReader(conn, upStreamDataChannel, upStreamControlChannel)

	downStreamDataChannel := make(chan []byte)
	downStreamControlChannel := make(chan string)

	go DownStreamReader(tconn, downStreamDataChannel, downStreamControlChannel)

	maxDuringTime := timeout_ms / 2

	intervalTime := maxDuringTime / 3

	ticker := time.NewTicker(time.Millisecond * time.Duration(intervalTime))

	flushCounter := 0

	relayData := &RelayData{
		Buff: make([]byte, 0),
		OK:   true,
	}

loop:
	for {

		select {

		case upbuffer := <-upStreamDataChannel:
			//tconn.Write([]byte(upbuffer))
			relayData.Buff = append(relayData.Buff, upbuffer...)

			//RelayUpStreamDataBySplit(relayData, tconn, segSize, false)
			hander(relayData, tconn, segSize, false)

			if !relayData.OK {

				tconn.Close()
				conn.Close()

				<-upStreamControlChannel

				<-downStreamControlChannel

				break loop
			}

		case <-upStreamControlChannel:

			tconn.Close()
			conn.Close()

			<-downStreamControlChannel

			break loop

		case downbuffer := <-downStreamDataChannel:
			conn.Write([]byte(downbuffer))

		case <-downStreamControlChannel:
			tconn.Close()
			conn.Close()

			<-upStreamControlChannel

			break loop

		case <-ticker.C:

			if flushCounter > 2 {

				flushCounter = 0

				//RelayUpStreamDataBySplit(relayData, tconn, segSize, true)
				hander(relayData, tconn, segSize, true)

				if !relayData.OK {

					tconn.Close()
					conn.Close()

					<-upStreamControlChannel

					<-downStreamControlChannel

					break loop
				}

			} else {

				flushCounter++

				//RelayUpStreamDataBySplit(relayData, tconn, segSize, false)
				hander(relayData, tconn, segSize, false)

				if !relayData.OK {

					tconn.Close()
					conn.Close()

					<-upStreamControlChannel

					<-downStreamControlChannel

					break loop
				}
			}

		}
	}
}

func usage() {
	fmt.Println("**************************************************************************************")
	fmt.Println(" ./msgMonkeyKing <listenPort> <dstAddr> <dstPort> <policy> <segmentSize> <timeout_ms>")
	fmt.Println("")
	fmt.Println("Note:")
	fmt.Println("\ttimeout_ms:{value >= 10}")
	fmt.Println("\tThe timeout_ms means the timeout value(millisecond) of your application (client to server)")
	fmt.Println("\tThe timeout_ms value will be used  as reference value to simulate the network delay and avoid your application timeout.")
	fmt.Println("")
	fmt.Println("\tpolicy:{split/merge}")
	fmt.Println("\tIt's used to simulate the tcp packet Fragmentation and Assembly.")
	fmt.Println("\tif policy=split, tha data will be splited(forword data size <= segmentSize)")
	fmt.Println("\tif policy=merge, tha data will be  merged(forword data size >= segmentSize)")
	fmt.Println("")
	os.Exit(0)
}

func getParams(args []string) (string, string, string, string, int, int) {

	listenPort := args[1]

	dstAddr := args[2]

	dstPort := args[3]

	policy := args[4]

	if !strings.EqualFold(policy, "split") && !strings.EqualFold(policy, "merge") {

		fmt.Println("\nError:policy must be 'split' or 'merge'")
		fmt.Println("input policy=", policy)

		usage()

		os.Exit(0)
	}

	segSize, err := strconv.Atoi(args[5])

	if err != nil || segSize <= 0 {
		fmt.Println("\nError:segSize param error")
		usage()
		os.Exit(0)
	}

	timeout_ms, err := strconv.Atoi(args[6])

	if err != nil || timeout_ms <= 10 {
		fmt.Println("\nError:timeout_ms param error")
		usage()
		os.Exit(0)
	}

	return listenPort, dstAddr, dstPort, policy, segSize, timeout_ms
}

func main() {

	if len(os.Args) < 7 {
		usage()
	}

	listenPort, dstAddr, dstPort, policy, segSize, timeout_ms := getParams(os.Args)

	LogMsg("Input parameters:", listenPort, dstAddr, dstPort, policy, segSize, timeout_ms)

	listener, error := net.Listen("tcp", ":"+listenPort)

	defer listener.Close()

	if error != nil {

		fmt.Println(error)

		return

	} else {

		for {

			LogMsg("Waiting for connections.....")

			connection, error := listener.Accept()

			if error != nil {

				LogMsg("error: ", error)

				return

			} else {

				if strings.EqualFold(policy, "split") {

					go Moneky(connection, dstAddr, dstPort, RelayUpStreamDataBySplit,
						segSize, timeout_ms)

				} else {

					go Moneky(connection, dstAddr, dstPort, RelayUpStreamDataByMerge,
						segSize, timeout_ms)

				}

			}
		}
	}

}
