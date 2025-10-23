package daemon

import (
	"fmt"
	"net"

	msgpacklib "github.com/vmihailenco/msgpack/v5"
	"github.com/woolawin/catalogue/internal"
)

type Client struct {
}

func NewClient() Client {
	return Client{}
}

func (client *Client) Send(cmd Cmd) (bool, any, error) {

	conn, err := net.Dial("unix", path)
	if err != nil {
		return false, nil, internal.ErrOf(err, "can not connect to socket")
	}
	defer conn.Close()

	msg := Message{Cmd: &cmd}

	writer := msgpacklib.NewEncoder(conn)
	reader := msgpacklib.NewDecoder(conn)

	err = writer.Encode(msg)
	if err != nil {
		return false, nil, internal.ErrOf(err, "can not message daemon")
	}

	for {
		reply := Message{}
		err = reader.Decode(&reply)
		if err != nil {
			return false, nil, internal.ErrOf(err, "can not read reply from daemon")
		}

		if msg.Log != nil {
			fmt.Println(msg.Log.Value)
		}

		if msg.End != nil {
			return msg.End.Ok, msg.End.Value, nil
		}
	}

}
