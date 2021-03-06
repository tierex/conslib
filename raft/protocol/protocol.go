package protocol

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Host struct {
	Domain string
	Port   int
}

func (host Host) String() string {
	return fmt.Sprintf("%s:%d", host.Domain, host.Port)
}

type RaftEventHandler interface {
	OnAppendEntriesReceived(args *AppendArgs, result *AppendResult) error
	OnRequestVoteReceived(args *RequestArgs, result *RequestResult) error
}

type Protocol interface {
	NewSender(host *Host) (Sender, error)
	RegisterListener(host *Host, eventHandler RaftEventHandler) error
}

func NewProtocol() Protocol {
	return new(RPCProtocol)
}

type RPCProtocol struct {
}

func (protocol *RPCProtocol) NewSender(host *Host) (Sender, error) {
	sender := new(RPCSender)
	client, err := rpc.DialHTTP("tcp", host.String())
	log.Println("Establishing connection with: ", host)
	if err != nil {
		log.Println("Warning: Cannot establish connection. ", err)
		return nil, nil
	}
	sender.RpcClient = client
	sender.protocol = protocol
	sender.destHost = host
	return sender, nil
}

func (protocol *RPCProtocol) RegisterListener(host *Host, eventHandler RaftEventHandler) error {
	receiver := &RPCReceiver{protocol, eventHandler}
	rpc.Register(receiver)
	rpc.HandleHTTP()
	fmt.Println("Start PRC litener", host.String())
	l, e := net.Listen("tcp", host.String())
	if e != nil {
		log.Fatal("Error while initialising RPC server:", e)
		return e
	}

	go http.Serve(l, nil)
	return nil
}
