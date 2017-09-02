package expose

import (
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/inu1255/expose/config"
	"golang.org/x/net/ipv4"
)

// forward r --> w
func Forward(r io.Reader, w io.Writer) {
	b := make([]byte, 1024)
	for {
		n, err := r.Read(b)
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
		_, err = w.Write(b[:n])
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
	}
}

// forward r --> w with timeout
func ForwardUDP(r *ioUDP, w io.Writer, timeout time.Duration) {
	b := make([]byte, 1024)
	for {
		if timeout > 0 {
			r.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err := r.Read(b)
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
		_, err = w.Write(b[:n])
		if err != nil {
			log.Println(reflect.TypeOf(r), err)
			return
		}
	}
}

// UDPConn --> fat interface io.WriteReader
type ioUDP struct {
	*net.UDPConn
	addr net.Addr
}

func (this *ioUDP) Read(b []byte) (int, error) {
	n, addr, err := this.UDPConn.ReadFrom(b)
	if err == nil && addr.String() != this.addr.String() {
		return 0, nil
	}
	return n, err
}

func (this *ioUDP) Write(b []byte) (int, error) {
	if os.Geteuid() == 0 {
		return WriteUdp(config.Cfg.Client.UdpAddress, this.addr.String(), b)
	}
	return this.UDPConn.WriteTo(b, this.addr)
}

// in --> conn
// conn --> out
func (conn *ioUDP) Tee(in io.Reader, out io.Writer) {
	go Forward(in, conn)
	Forward(conn, out)
}

func NewIoUDP(conn *net.UDPConn, addr net.Addr) *ioUDP {
	return &ioUDP{conn, addr}
}

func checkSum(msg []byte) uint16 {
	sum := 0
	for n := 1; n < len(msg)-1; n += 2 {
		sum += int(msg[n])*256 + int(msg[n+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var ans = uint16(^sum)
	return ans
}

func WriteUdp(src, dst string, buff []byte) (int, error) {
	//目的IP
	dstAddr, _ := net.ResolveUDPAddr("udp", dst)
	dstIP := dstAddr.IP
	//源IP
	srcAddr, _ := net.ResolveUDPAddr("udp", src)
	srcIP := srcAddr.IP
	//填充ip首部
	iph := &ipv4.Header{
		Version: ipv4.Version,
		//IP头长一般是20
		Len: ipv4.HeaderLen,
		TOS: 0x00,
		//buff为数据
		TotalLen: ipv4.HeaderLen + 8 + len(buff),
		TTL:      64,
		Flags:    ipv4.DontFragment,
		FragOff:  0,
		Protocol: 17,
		Checksum: 0,
		Src:      srcIP,
		Dst:      dstIP,
	}

	h, err := iph.Marshal()
	if err != nil {
		log.Fatalln(err)
	}
	//计算IP头部校验值
	iph.Checksum = int(checkSum(h))
	//填充udp首部
	//udp伪首部
	udph := make([]byte, 20)
	//源ip地址
	udph[0], udph[1], udph[2], udph[3] = srcIP[12], srcIP[13], srcIP[14], srcIP[15]
	//目的ip地址
	udph[4], udph[5], udph[6], udph[7] = dstIP[12], dstIP[13], dstIP[14], dstIP[15]
	//协议类型
	udph[8], udph[9] = 0x00, 0x11
	//udp头长度
	udph[10], udph[11] = 0x00, byte(len(buff)+8)
	//下面开始就真正的udp头部
	//源端口号
	udph[12], udph[13] = byte(srcAddr.Port>>8&255), byte(srcAddr.Port&255)
	//目的端口号
	udph[14], udph[15] = byte(dstAddr.Port>>8&255), byte(dstAddr.Port&255)
	//udp头长度
	n := len(buff) + 8
	udph[16], udph[17] = byte(n>>8&255), byte(n&255)
	//校验和
	udph[18], udph[19] = 0x00, 0x00
	//计算校验值
	check := checkSum(append(udph, buff...))
	udph[18], udph[19] = byte(check>>8&255), byte(check&255)
	listener, err := net.ListenPacket("ip4:udp", "0.0.0.0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	//listener 实现了net.PacketConn接口
	r, err := ipv4.NewRawConn(listener)
	if err != nil {
		log.Fatal(err)
	}

	//发送自己构造的UDP包
	return len(buff), r.WriteTo(iph, append(udph[12:20], buff...), nil)
}
