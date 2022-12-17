package TCP

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// 랜덤 포트에 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// 병렬로 실행 중인 다른 함수가 끝날 때까지 대기시키기 위해 chan 사용, struct{}형이 Go의 자료형 중 가장 공간을 작게 차지하기 때문에 struct{} 형을 받게 한다.
	done := make(chan struct{})
	go func() { // 리스너를 고루틴에서 시작해서 이후의 테스트에서 클라이언트 측에서 연결할 수 있도록한다.
		defer func() { done <- struct{}{} }()

		for {
			/*
				리스너의 Accept 메서드는 리스너가 종료되면, 즉시 블로킹이 해제되고 에러를 반환한다.
				이 때 에러는 무언가 실패했다는 의미가 아니다.
			*/
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return // 리스너의 고루틴이 종료되며 테스트가 완료된다.
			}

			go func(c net.Conn) {
				defer func() {
					c.Close() // 커넥션 핸들러는 연결 객체의 close 메서드를 호출하며 종료, close 메서드는 FIN 패킷을 전송하며 TCP 의 graceful close 를 마무리
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf) // FIN 패킷을 받고 나면 Read 메서드는 io.EOF 에러를 반환, 리스너 측에서는 반대편 연결이 종료되었다는 의미
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("reveived: %q", buf[:n])
				}
			}(conn)
		}
	}()

	/*
		net.Dial 함수는 tcp 같은 네트워크 종류와 IP 주소, 포트의 조합을 매개변수로 받는다.
		Dial 함수에서 두 번째 매개변수로 받은 IP 주소, 포트를 이용해 리스너로 연결을 시도.
		IP 주소 대신 호스트명을 사용할 수도 있고, 포트 대신에 http 와 같은 서비스명을 사용할 수도 있다.
		호스트명이 하나 이상의 주소로 해석되면 go는 성공할 때까지 각각의 IP 주소에 연결을 시도한다.
	*/
	conn, err := net.Dial("tcp", listener.Addr().String()) // Dial 함수는 연결 객체와 에러 인터페이스의 값을 반환.
	if err != nil {
		t.Fatal(err)
	}

	conn.Close() // 리스너 연결을 성골적으로 수립한 후, 클라이언트 측에서 graceful close 를 시작한다.
	<-done
	listener.Close() // 마지막으로 리스너를 종료
	<-done
}
