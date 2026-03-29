package domain

import (
	"io"
	"log"
	"os"
	"time"
)

// 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// 日志配置
type LogConfig struct {
	Level   LogLevel
	Output  io.Writer
	Prefix  string
	Enabled bool
}

// 默认日志配置
var defaultLogConfig = LogConfig{
	Level:   LogLevelInfo,
	Output:  os.Stdout,
	Prefix:  "[FSM] ",
	Enabled: true,
}

// 状态机接口
type StateMachine interface {
	CurrentState() string
	HandleEvent(event string, args ...interface{}) error
	CanHandleEvent(event string) bool
}

// 状态机实现
type FSM struct {
	currentState string
	states       map[string]bool
	events       map[string]bool
	transitions  map[string]map[string]string // fromState -> event -> toState
	callbacks    map[string]func(*Event)
	logConfig    *LogConfig
	logger       *log.Logger
}

// 事件结构
type Event struct {
	Name      string
	Src       string
	Dst       string
	Args      []interface{}
	Timestamp int64
}

// 状态机选项
type FSMOption func(*FSM)

// 日志选项
func WithLogConfig(config LogConfig) FSMOption {
	return func(fsm *FSM) {
		fsm.logConfig = &config
		fsm.logger = log.New(config.Output, config.Prefix, log.LstdFlags|log.Lmicroseconds)
	}
}

// 启用日志
func WithLogging(enabled bool) FSMOption {
	return func(fsm *FSM) {
		if fsm.logConfig == nil {
			fsm.logConfig = &defaultLogConfig
			fsm.logger = log.New(defaultLogConfig.Output, defaultLogConfig.Prefix, log.LstdFlags|log.Lmicroseconds)
		}
		fsm.logConfig.Enabled = enabled
	}
}

// 新状态机
func NewFSM(initialState string, options ...FSMOption) *FSM {
	fsm := &FSM{
		currentState: initialState,
		states:       make(map[string]bool),
		events:       make(map[string]bool),
		transitions:  make(map[string]map[string]string),
		callbacks:    make(map[string]func(*Event)),
		logConfig:    &defaultLogConfig,
		logger:       log.New(defaultLogConfig.Output, defaultLogConfig.Prefix, log.LstdFlags|log.Lmicroseconds),
	}

	// 添加初始状态
	fsm.states[initialState] = true

	// 应用选项
	for _, option := range options {
		option(fsm)
	}

	// 记录状态机初始化
	fsm.logInfo("State machine initialized with state: %s", initialState)

	return fsm
}

// 状态选项
func WithStates(states ...string) FSMOption {
	return func(fsm *FSM) {
		for _, state := range states {
			fsm.states[state] = true
		}
	}
}

// 事件选项
func WithEvents(events ...string) FSMOption {
	return func(fsm *FSM) {
		for _, event := range events {
			fsm.events[event] = true
		}
	}
}

// 转换选项
func WithTransition(fromState, event, toState string) FSMOption {
	return func(fsm *FSM) {
		if fsm.transitions[fromState] == nil {
			fsm.transitions[fromState] = make(map[string]string)
		}
		fsm.transitions[fromState][event] = toState
		fsm.states[fromState] = true
		fsm.states[toState] = true
		fsm.events[event] = true
	}
}

// 回调选项
func WithCallback(event string, callback func(*Event)) FSMOption {
	return func(fsm *FSM) {
		fsm.callbacks[event] = callback
	}
}

// 当前状态
func (fsm *FSM) CurrentState() string {
	return fsm.currentState
}

// 日志方法
func (fsm *FSM) logDebug(format string, args ...interface{}) {
	if fsm.logConfig.Enabled && fsm.logConfig.Level <= LogLevelDebug && fsm.logger != nil {
		fsm.logger.Printf("[DEBUG] "+format, args...)
	}
}

func (fsm *FSM) logInfo(format string, args ...interface{}) {
	if fsm.logConfig.Enabled && fsm.logConfig.Level <= LogLevelInfo && fsm.logger != nil {
		fsm.logger.Printf("[INFO] "+format, args...)
	}
}

func (fsm *FSM) logWarning(format string, args ...interface{}) {
	if fsm.logConfig.Enabled && fsm.logConfig.Level <= LogLevelWarning && fsm.logger != nil {
		fsm.logger.Printf("[WARN] "+format, args...)
	}
}

func (fsm *FSM) logError(format string, args ...interface{}) {
	if fsm.logConfig.Enabled && fsm.logConfig.Level <= LogLevelError && fsm.logger != nil {
		fsm.logger.Printf("[ERROR] "+format, args...)
	}
}

// 处理事件
func (fsm *FSM) HandleEvent(event string, args ...interface{}) error {
	fsm.logDebug("Received event: %s, current state: %s", event, fsm.currentState)

	if !fsm.events[event] {
		fsm.logWarning("Ignoring unknown event: %s", event)
		return nil
	}

	toState, ok := fsm.transitions[fsm.currentState][event]
	if !ok {
		fsm.logWarning("Cannot transition from state %s with event %s", fsm.currentState, event)
		return nil
	}

	// 创建事件
	e := &Event{
		Name:      event,
		Src:       fsm.currentState,
		Dst:       toState,
		Args:      args,
		Timestamp: time.Now().UnixNano(),
	}

	// 记录状态转换
	fsm.logInfo("State transition: %s -> %s (event: %s, args: %v)",
		fsm.currentState, toState, event, args)

	// 执行回调
	if callback, ok := fsm.callbacks[event]; ok {
		fsm.logDebug("Executing callback for event: %s", event)
		callback(e)
	}

	// 更新状态
	fsm.currentState = toState

	fsm.logDebug("State updated to: %s", fsm.currentState)

	return nil
}

// 是否可以处理事件
func (fsm *FSM) CanHandleEvent(event string) bool {
	_, ok := fsm.transitions[fsm.currentState][event]
	return ok
}

// 机器人状态机
func NewBotFSM() *FSM {
	return NewFSM("idle",
		WithStates("idle", "processing", "error"),
		WithEvents("assign", "complete", "error", "recover"),
		WithTransition("idle", "assign", "processing"),
		WithTransition("processing", "complete", "idle"),
		WithTransition("processing", "error", "error"),
		WithTransition("error", "recover", "idle"),
	)
}

// 订单状态机
func NewOrderFSM() *FSM {
	return NewFSM("pending",
		WithStates("pending", "processing", "complete", "cancelled", "error"),
		WithEvents("assign", "complete", "cancel", "fail", "retry"),
		WithTransition("pending", "assign", "processing"),
		WithTransition("processing", "complete", "complete"),
		WithTransition("pending", "cancel", "cancelled"),
		WithTransition("processing", "cancel", "cancelled"),
		WithTransition("processing", "fail", "error"),
		WithTransition("error", "retry", "pending"),
	)
}
