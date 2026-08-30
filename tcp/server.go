package tcp

/**
 * A tcp server
 */

import (
	"context"
	"fmt"
	"go-redis/interface/tcp"
	"go-redis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Config stores tcp server properties
type Config struct {
    Address string
}

// ListenAndServeWithSignal binds port and handle requests, blocking until receive stop signal
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
    closeChan := make(chan struct{})
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
    defer signal.Stop(sigCh)

    go func() {
        sig := <-sigCh
        switch sig {
        case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
            select {
            case <-closeChan:
            default:
                close(closeChan)
            }
        }
    }()

    listener, err := net.Listen("tcp", cfg.Address)
    if err != nil {
        return err
    }
    logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
    ListenAndServe(listener, handler, closeChan)
    return nil
}

// ListenAndServe binds port and handle requests, blocking until close
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
    var shutdownOnce sync.Once
    shutdown := func() {
        shutdownOnce.Do(func() {
            logger.Info("shutting down...")
            _ = listener.Close() // listener.Accept() will return err immediately
            _ = handler.Close()  // close connections
        })
    }

    // listen signal
    go func() {
        <-closeChan
        shutdown()
    }()

    // listen port
    defer shutdown()
    ctx := context.Background()
    var waitDone sync.WaitGroup
    for {
        conn, err := listener.Accept()
        if err != nil {
            break
        }
        // handle
        logger.Info("accept link")
        waitDone.Add(1)
        go func() {
            defer func() {
                waitDone.Done()
            }()
            handler.Handle(ctx, conn)
        }()
    }
    waitDone.Wait()
}
