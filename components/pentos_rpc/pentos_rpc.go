package main

import (
	"context"
	"encoding/json"
	"github.com/johanliu/essos"
	"github.com/johanliu/essos/components"
	"google.golang.org/grpc"

	pb "./grpc"
	"github.com/labstack/gommon/log"
)

const (
	VERSION = "1.0"
)

type pentos_rpc struct {
	ops map[string]essos.Operation
}

var _ip string
var _port string
var conn grpc.ClientConn
var client pb.ServiceClient

func (p *pentos_rpc) Discover() map[string]essos.Operation {
	return p.ops
}

func init() {
	components.Add("pipeline",
		&pentos_rpc{
			ops: map[string]essos.Operation{
				"addTerran": addTerran("addTerran"),
				"deleteTerran":   deleteTerran("deleteTerran"),
				"listTerrans": listTerrans("listTerrans"),
				"checkComplete": checkComplete("checkComplete"),
				"listIncompleteFlags": listIncompleteFlags("listIncompleteFlags"),
				"markComplete": markComplete("markComplete"),
			},
		})
}

func (p *pentos_rpc)InitConnection(ip string, port string) (error) {
	_ip = ip
	_port = port
	return connect()
}

func connect() (error) {
	log.Infof("Rpc init connection to %s:%s", _ip, _port)
	conn, err := grpc.Dial(_ip + ":" + _port, grpc.WithInsecure())
	if err != nil {
		log.Warnf("Can't connect to : %v", err)
		return err
	}
	client = pb.NewServiceClient(conn)
	return nil
}

func (p *pentos_rpc)stop() {
	conn.Close()
}

type addTerran string

func (addTerran) Description() string {
	return "add terran flag"
}

func (obj addTerran) Do(ctx context.Context, args []string) (context.Context, error) {
	reply, err := client.AddTerran(context.Background(), &pb.Request{Arg:args})
	return result(ctx, reply, err, obj.Description())
}

type deleteTerran string

func (deleteTerran) Description() string {
	return "delete terran flag"
}

func (obj deleteTerran) Do(ctx context.Context, args []string) (context.Context, error) {
	reply, err := client.DeleteTerran(context.Background(), &pb.Request{Arg:args})
	return result(ctx, reply, err, obj.Description())
}

type listTerrans string

func (listTerrans) Description() string {
	return "list terran flags"
}

func (obj listTerrans) Do(ctx context.Context, args []string) (context.Context, error) {
	reply, err := client.ListTerrans(context.Background(), &pb.Request{Arg:args})
	return result(ctx, reply, err, obj.Description())
}

type checkComplete string

func (checkComplete) Description() string {
	return "check timestamp complete"
}

func (obj checkComplete) Do(ctx context.Context, args []string) (context.Context, error) {
	reply, err := client.CheckComplete(context.Background(), &pb.Request{Arg:args})
	return result(ctx, reply, err, obj.Description())
}

type listIncompleteFlags string

func (listIncompleteFlags ) Description() string {
	return "list incomplete terran flags in minute"
}

func (obj listIncompleteFlags ) Do(ctx context.Context, args []string) (context.Context, error) {
	reply, err := client.ListIncompleteFlags(context.Background(), &pb.Request{Arg:args})
	return result(ctx, reply, err, obj.Description())
}

type markComplete string

func (markComplete) Description() string {
	return "mark complete flag at minute"
}

func (obj markComplete) Do(ctx context.Context, args []string) (context.Context, error) {
	reply, err := client.MarkComplete(context.Background(), &pb.Request{Arg:args})
	return result(ctx, reply, err, obj.Description())
}

func result(ctx context.Context, reply *pb.Reply, err error, description string) (context.Context, error) {
	var code int
	var msg string

	if err != nil {
		code = 503
		msg = "failed to do rpc call: " + description + " " + err.Error()
		log.Warn(msg)
	} else {
		content, err := json.Marshal(reply.Msg);
		if err == nil {
			code = 200
			msg = string(content)
		} else {
			code = 503
			msg = "failed to do rpc call: " + description + " " + err.Error()
			log.Warn(msg)
		}
	}

	result := essos.Response{
		Code:    code,
		Message: []byte(msg),
	}
	ctx = context.WithValue(ctx, "result", result)
	return ctx, nil
}

