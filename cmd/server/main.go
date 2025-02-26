package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// gRPC 서버를 실행할 포트 설정
	const port = ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("포트 %s 에서 리스닝 실패: %v", port, err)
	}

	// gRPC 서버 생성 (필요한 인터셉터나 옵션 추가 가능)
	grpcServer := grpc.NewServer()

	// TODO: 실제 AuthService 구현체를 등록합니다.
	// 예시:
	// authpb.RegisterAuthServiceServer(grpcServer, grpc.NewAuthServiceServer(...))

	// gRPC 서버를 별도의 고루틴에서 실행
	go func() {
		log.Printf("gRPC 서버가 포트 %s 에서 시작되었습니다.", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC 서버 실행 중 오류 발생: %v", err)
		}
	}()

	// OS 신호(Interrupt, SIGTERM) 수신을 위한 채널 생성
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 신호가 올 때까지 대기

	log.Println("gRPC 서버 종료 준비 중...")

	// graceful shutdown: 기존 연결들을 마무리할 시간을 제공
	gracefulStop(grpcServer, 5*time.Second)

	log.Println("gRPC 서버가 정상적으로 종료되었습니다.")
}

// gracefulStop는 지정된 시간 동안 기존 연결을 종료시키고, 서버를 안전하게 중지합니다.
func gracefulStop(server *grpc.Server, timeout time.Duration) {
	// gracefulStop은 blocking 함수로, 기존 연결들을 처리한 후 종료합니다.
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	// 지정된 시간 내에 종료되지 않으면 강제 종료
	select {
	case <-done:
		// 정상적으로 종료됨
	case <-time.After(timeout):
		log.Println("graceful shutdown 시간 초과, 강제 종료합니다.")
		server.Stop()
	}
}
