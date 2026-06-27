package task

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

const (
	DebounceWindow = 15 * time.Second
	TaskTimeout    = 90 * time.Second
	MaxRetry       = 5 // after this the task is archived (asynq's dead-letter)
	Concurrency    = 32
	QueueName      = "default"
	ShutdownGrace  = 30 * time.Second
)

type Client struct {
	inner *asynq.Client
	log   *slog.Logger
}

func NewClient(redisOpt asynq.RedisConnOpt, log *slog.Logger) *Client {
	return &Client{inner: asynq.NewClient(redisOpt), log: log}
}

func (c *Client) Close() error { return c.inner.Close() }

func (c *Client) EnqueueReply(ctx context.Context, businessID, conversationID string) error {
	task, err := NewChatReplyTask(businessID, conversationID)
	if err != nil {
		return err
	}
	_, err = c.inner.EnqueueContext(ctx, task,
		asynq.TaskID(DebounceTaskID(conversationID)),
		asynq.ProcessIn(DebounceWindow),
		asynq.MaxRetry(MaxRetry),
		asynq.Timeout(TaskTimeout),
		asynq.Queue(QueueName),
	)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		c.log.Debug("reply already scheduled, coalescing", "conversation_id", conversationID)
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) EnqueueReplyDrain(ctx context.Context, businessID, conversationID string) error {
	task, err := NewChatReplyTask(businessID, conversationID)
	if err != nil {
		return err
	}
	_, err = c.inner.EnqueueContext(ctx, task,
		asynq.ProcessIn(DebounceWindow),
		asynq.MaxRetry(MaxRetry),
		asynq.Timeout(TaskTimeout),
		asynq.Queue(QueueName),
	)
	return err
}

type Server struct {
	srv *asynq.Server
	mux *asynq.ServeMux
	log *slog.Logger
}

func NewServer(redisOpt asynq.RedisConnOpt, log *slog.Logger) *Server {
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency:     Concurrency,
		Queues:          map[string]int{QueueName: 1},
		ShutdownTimeout: ShutdownGrace,
		Logger:          &slogAdapter{log: log},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			lvl := slog.LevelWarn
			if retried >= maxRetry {
				lvl = slog.LevelError
			}
			log.Log(ctx, lvl, "task failed",
				"type", task.Type(), "retried", retried, "max_retry", maxRetry, "err", err)
		}),
	})
	return &Server{srv: srv, mux: asynq.NewServeMux(), log: log}
}

func (s *Server) Register(taskType string, h func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(taskType, h)
}

func (s *Server) Run() error {
	s.log.Info("asynq server starting", "concurrency", Concurrency, "queue", QueueName)
	return s.srv.Run(s.mux)
}

func (s *Server) Start() error {
	s.log.Info("asynq server starting", "concurrency", Concurrency, "queue", QueueName)
	return s.srv.Start(s.mux)
}

func (s *Server) Shutdown() {
	s.log.Info("asynq server shutting down")
	s.srv.Shutdown()
}

type slogAdapter struct{ log *slog.Logger }

func (a *slogAdapter) Debug(args ...interface{}) { a.log.Debug(sprint(args)) }
func (a *slogAdapter) Info(args ...interface{})  { a.log.Info(sprint(args)) }
func (a *slogAdapter) Warn(args ...interface{})  { a.log.Warn(sprint(args)) }
func (a *slogAdapter) Error(args ...interface{}) { a.log.Error(sprint(args)) }
func (a *slogAdapter) Fatal(args ...interface{}) { a.log.Error(sprint(args), "fatal", true) }

func sprint(args []interface{}) string {
	if len(args) == 1 {
		if s, ok := args[0].(string); ok {
			return s
		}
	}
	return fmt.Sprint(args...)
}
