package main

import (
	"bufio"
	"fmt"
	"mcdonalds-order-controller/domain"
	"os"
)

func main() {
	fmt.Println("=== 状态机日志追踪功能演示 ===\n")

	// 演示 1: 控制台输出（默认）
	fmt.Println("【演示 1】控制台日志输出")
	fmt.Println("-----------------------------")
	demoConsoleLogging()
	fmt.Println()

	// 演示 2: 文件输出
	fmt.Println("【演示 2】文件日志输出")
	fmt.Println("-----------------------------")
	demoFileLogging()
	fmt.Println()

	// 演示 3: 不同日志级别
	fmt.Println("【演示 3】日志级别过滤")
	fmt.Println("-----------------------------")
	demoLogLevels()
	fmt.Println()

	fmt.Println("=== 演示完成 ===")
}

func demoConsoleLogging() {
	// 创建机器人状态机，使用默认配置（控制台输出）
	botFSM := domain.NewBotFSM()
	
	fmt.Println("▶ 机器人状态机初始化完成，状态:", botFSM.CurrentState())
	
	// 触发状态转换
	fmt.Println("▶ 触发 'assign' 事件...")
	botFSM.HandleEvent("assign", "Order #123", "VIP")
	
	fmt.Println("▶ 触发 'complete' 事件...")
	botFSM.HandleEvent("complete")
	
	fmt.Println("▶ 最终状态:", botFSM.CurrentState())
}

func demoFileLogging() {
	// 创建日志文件
	logFile, err := os.Create("fsm_demo.log")
	if err != nil {
		fmt.Printf("❌ 创建日志文件失败: %v\n", err)
		return
	}
	defer logFile.Close()
	
	// 创建订单状态机，使用文件输出
	logConfig := domain.LogConfig{
		Level:   domain.LogLevelDebug,
		Output:  logFile,
		Prefix:  "[Demo Order FSM] ",
		Enabled: true,
	}
	
	orderFSM := domain.NewFSM("pending",
		domain.WithStates("pending", "processing", "complete", "cancelled", "error"),
		domain.WithEvents("assign", "complete", "cancel", "fail", "retry"),
		domain.WithTransition("pending", "assign", "processing"),
		domain.WithTransition("processing", "complete", "complete"),
		domain.WithTransition("pending", "cancel", "cancelled"),
		domain.WithTransition("processing", "cancel", "cancelled"),
		domain.WithTransition("processing", "fail", "error"),
		domain.WithTransition("error", "retry", "pending"),
		domain.WithLogConfig(logConfig),
	)
	
	// 触发一些状态转换
	orderFSM.HandleEvent("assign")
	orderFSM.HandleEvent("complete")
	
	// 读取并显示日志文件内容
	logFile.Seek(0, 0)
	scanner := bufio.NewScanner(logFile)
	fmt.Println("📄 日志文件内容:")
	for scanner.Scan() {
		fmt.Println("   ", scanner.Text())
	}
	
	fmt.Println("✅ 日志已写入到 fsm_demo.log 文件")
}

func demoLogLevels() {
	// 测试 INFO 级别（只显示 INFO 及以上）
	fmt.Println("1️⃣  使用 LogLevelInfo:")
	var infoBuf []byte
	infoLogConfig := domain.LogConfig{
		Level:   domain.LogLevelInfo,
		Output:  &infoBuffer{&infoBuf},
		Prefix:  "[INFO Test] ",
		Enabled: true,
	}
	
	infoFSM := domain.NewFSM("a",
		domain.WithStates("a", "b"),
		domain.WithEvents("to_b"),
		domain.WithTransition("a", "to_b", "b"),
		domain.WithLogConfig(infoLogConfig),
	)
	
	infoFSM.HandleEvent("to_b")
	fmt.Println("   输出:", string(infoBuf))
	
	// 测试 ERROR 级别（只显示 ERROR）
	fmt.Println("\n2️⃣  使用 LogLevelError:")
	var errBuf []byte
	errLogConfig := domain.LogConfig{
		Level:   domain.LogLevelError,
		Output:  &infoBuffer{&errBuf},
		Prefix:  "[ERROR Test] ",
		Enabled: true,
	}
	
	errFSM := domain.NewFSM("x",
		domain.WithStates("x", "y"),
		domain.WithEvents("to_y"),
		domain.WithTransition("x", "to_y", "y"),
		domain.WithLogConfig(errLogConfig),
	)
	
	errFSM.HandleEvent("to_y")
	if len(errBuf) == 0 {
		fmt.Println("   （无输出，因为 INFO 级别日志被过滤）")
	} else {
		fmt.Println("   输出:", string(errBuf))
	}
	
	// 测试禁用日志
	fmt.Println("\n3️⃣  禁用日志:")
	var disabledBuf []byte
	disabledLogConfig := domain.LogConfig{
		Level:   domain.LogLevelDebug,
		Output:  &infoBuffer{&disabledBuf},
		Prefix:  "[Disabled Test] ",
		Enabled: false,
	}
	
	disabledFSM := domain.NewFSM("start",
		domain.WithStates("start", "end"),
		domain.WithEvents("finish"),
		domain.WithTransition("start", "finish", "end"),
		domain.WithLogConfig(disabledLogConfig),
	)
	
	disabledFSM.HandleEvent("finish")
	if len(disabledBuf) == 0 {
		fmt.Println("   （无输出，日志已禁用）")
	}
}

// 简单的缓冲区实现，用于测试
type infoBuffer struct {
	buf *[]byte
}

func (ib *infoBuffer) Write(p []byte) (n int, err error) {
	*ib.buf = append(*ib.buf, p...)
	return len(p), nil
}