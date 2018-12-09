package network

import (
	"context"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"net"
)

type udpHandler struct {
	conf *ServerConf
	ts   *Server

	conn      *net.UDPConn
	numInvoke int32
}

func (h *udpHandler) Listen() error {
	cfg := h.conf
	addr, err := net.ResolveUDPAddr("udp4", cfg.Address)
	if err != nil {
		return err
	}
	h.conn, err = net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	logger.Info("UDP listen", h.conn.LocalAddr())
	return nil
}

func (h *udpHandler) Handle() error {
	buffer := make([]byte, 65535)
	for !h.ts.isClosed {
		n, udpAddr, err := h.conn.ReadFromUDP(buffer)
		if err != nil {
			if isNoDataError(err) {
				continue
			} else {
				logger.Error(fmt.Sprintf("Close connection %s: %v", h.conf.Address, err))
				return err // TODO: check if necessary
			}
		}
		pkg := make([]byte, n)
		copy(pkg, buffer[0:n])
		go func() {
			ctx := context.Background()
			rsp := h.ts.invoke(ctx, pkg[4:]) // no need to check package
			if _, err := h.conn.WriteToUDP(rsp, udpAddr); err != nil {
				logger.Error(fmt.Sprintf("send pkg to %v failed %v", udpAddr, err))
			}
		}()
	}
	return nil
}
