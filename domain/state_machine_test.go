package domain

import (
	"testing"
)

func TestBotStateMachineTransitions(t *testing.T) {
	// 创建机器人状态机
	fsm := NewBotFSM()

	// 测试初始状态
	if fsm.CurrentState() != "idle" {
		t.Errorf("expected initial state to be 'idle', got %s", fsm.CurrentState())
	}

	// 测试从 idle 到 processing
	err := fsm.HandleEvent("assign")
	if err != nil {
		t.Errorf("expected no error when handling 'assign' event, got %v", err)
	}
	if fsm.CurrentState() != "processing" {
		t.Errorf("expected state to be 'processing' after 'assign' event, got %s", fsm.CurrentState())
	}

	// 测试从 processing 到 idle
	err = fsm.HandleEvent("complete")
	if err != nil {
		t.Errorf("expected no error when handling 'complete' event, got %v", err)
	}
	if fsm.CurrentState() != "idle" {
		t.Errorf("expected state to be 'idle' after 'complete' event, got %s", fsm.CurrentState())
	}

	// 测试从 idle 到 processing 再到 error
	err = fsm.HandleEvent("assign")
	if err != nil {
		t.Errorf("expected no error when handling 'assign' event, got %v", err)
	}
	if fsm.CurrentState() != "processing" {
		t.Errorf("expected state to be 'processing' after 'assign' event, got %s", fsm.CurrentState())
	}

	err = fsm.HandleEvent("error")
	if err != nil {
		t.Errorf("expected no error when handling 'error' event, got %v", err)
	}
	if fsm.CurrentState() != "error" {
		t.Errorf("expected state to be 'error' after 'error' event, got %s", fsm.CurrentState())
	}

	// 测试从 error 到 idle
	err = fsm.HandleEvent("recover")
	if err != nil {
		t.Errorf("expected no error when handling 'recover' event, got %v", err)
	}
	if fsm.CurrentState() != "idle" {
		t.Errorf("expected state to be 'idle' after 'recover' event, got %s", fsm.CurrentState())
	}

	// 测试无效事件
	err = fsm.HandleEvent("invalid_event")
	if err != nil {
		t.Errorf("expected no error when handling invalid event, got %v", err)
	}
	if fsm.CurrentState() != "idle" {
		t.Errorf("expected state to remain 'idle' after invalid event, got %s", fsm.CurrentState())
	}
}

func TestOrderStateMachineTransitions(t *testing.T) {
	// 创建订单状态机
	fsm := NewOrderFSM()

	// 测试初始状态
	if fsm.CurrentState() != "pending" {
		t.Errorf("expected initial state to be 'pending', got %s", fsm.CurrentState())
	}

	// 测试从 pending 到 processing
	err := fsm.HandleEvent("assign")
	if err != nil {
		t.Errorf("expected no error when handling 'assign' event, got %v", err)
	}
	if fsm.CurrentState() != "processing" {
		t.Errorf("expected state to be 'processing' after 'assign' event, got %s", fsm.CurrentState())
	}

	// 测试从 processing 到 complete
	err = fsm.HandleEvent("complete")
	if err != nil {
		t.Errorf("expected no error when handling 'complete' event, got %v", err)
	}
	if fsm.CurrentState() != "complete" {
		t.Errorf("expected state to be 'complete' after 'complete' event, got %s", fsm.CurrentState())
	}

	// 测试从 pending 到 cancelled
	fsm2 := NewOrderFSM()
	err = fsm2.HandleEvent("cancel")
	if err != nil {
		t.Errorf("expected no error when handling 'cancel' event, got %v", err)
	}
	if fsm2.CurrentState() != "cancelled" {
		t.Errorf("expected state to be 'cancelled' after 'cancel' event, got %s", fsm2.CurrentState())
	}

	// 测试从 processing 到 cancelled
	fsm3 := NewOrderFSM()
	err = fsm3.HandleEvent("assign")
	if err != nil {
		t.Errorf("expected no error when handling 'assign' event, got %v", err)
	}
	err = fsm3.HandleEvent("cancel")
	if err != nil {
		t.Errorf("expected no error when handling 'cancel' event, got %v", err)
	}
	if fsm3.CurrentState() != "cancelled" {
		t.Errorf("expected state to be 'cancelled' after 'cancel' event, got %s", fsm3.CurrentState())
	}

	// 测试从 processing 到 error 再到 pending
	fsm4 := NewOrderFSM()
	err = fsm4.HandleEvent("assign")
	if err != nil {
		t.Errorf("expected no error when handling 'assign' event, got %v", err)
	}
	err = fsm4.HandleEvent("fail")
	if err != nil {
		t.Errorf("expected no error when handling 'fail' event, got %v", err)
	}
	if fsm4.CurrentState() != "error" {
		t.Errorf("expected state to be 'error' after 'fail' event, got %s", fsm4.CurrentState())
	}

	err = fsm4.HandleEvent("retry")
	if err != nil {
		t.Errorf("expected no error when handling 'retry' event, got %v", err)
	}
	if fsm4.CurrentState() != "pending" {
		t.Errorf("expected state to be 'pending' after 'retry' event, got %s", fsm4.CurrentState())
	}

	// 测试无效事件
	err = fsm4.HandleEvent("invalid_event")
	if err != nil {
		t.Errorf("expected no error when handling invalid event, got %v", err)
	}
	if fsm4.CurrentState() != "pending" {
		t.Errorf("expected state to remain 'pending' after invalid event, got %s", fsm4.CurrentState())
	}
}

func TestBotStateMachineIntegration(t *testing.T) {
	// 创建机器人
	bot := NewBot(1)

	// 测试初始状态
	if bot.Status != Idle {
		t.Errorf("expected initial bot status to be Idle, got %v", bot.Status)
	}
	if bot.StateMachine.CurrentState() != "idle" {
		t.Errorf("expected initial state machine state to be 'idle', got %s", bot.StateMachine.CurrentState())
	}

	// 创建订单
	order := NewOrder(1, Normal)

	// 测试开始处理订单
	err := bot.StartProcessing(order)
	if err != nil {
		t.Errorf("expected no error when starting processing, got %v", err)
	}
	if bot.Status != Processing {
		t.Errorf("expected bot status to be Processing, got %v", bot.Status)
	}
	if bot.StateMachine.CurrentState() != "processing" {
		t.Errorf("expected state machine state to be 'processing', got %s", bot.StateMachine.CurrentState())
	}

	// 测试完成处理订单
	completedOrder := bot.CompleteProcessing()
	if completedOrder == nil {
		t.Error("expected completed order to not be nil")
	}
	if bot.Status != Idle {
		t.Errorf("expected bot status to be Idle, got %v", bot.Status)
	}
	if bot.StateMachine.CurrentState() != "idle" {
		t.Errorf("expected state machine state to be 'idle', got %s", bot.StateMachine.CurrentState())
	}
}

func TestOrderStateMachineIntegration(t *testing.T) {
	// 创建订单
	order := NewOrder(1, Normal)

	// 测试初始状态
	if order.Status != OrderPending {
		t.Errorf("expected initial order status to be OrderPending, got %v", order.Status)
	}
	if order.StateMachine.CurrentState() != "pending" {
		t.Errorf("expected initial state machine state to be 'pending', got %s", order.StateMachine.CurrentState())
	}

	// 测试标记为处理中
	order.MarkProcessing()
	if order.Status != OrderProcessing {
		t.Errorf("expected order status to be OrderProcessing, got %v", order.Status)
	}
	if order.StateMachine.CurrentState() != "processing" {
		t.Errorf("expected state machine state to be 'processing', got %s", order.StateMachine.CurrentState())
	}

	// 测试标记为完成
	order.MarkComplete()
	if order.Status != OrderComplete {
		t.Errorf("expected order status to be OrderComplete, got %v", order.Status)
	}
	if order.StateMachine.CurrentState() != "complete" {
		t.Errorf("expected state machine state to be 'complete', got %s", order.StateMachine.CurrentState())
	}
}