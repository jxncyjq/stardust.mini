package proto

import (
	"google.golang.org/protobuf/encoding/protowire"
)

// User protobuf消息
type User struct {
	UserName string
	Password string
}

func (x *User) Reset() {
	*x = User{}
}

func (x *User) GetUserName() string {
	if x != nil {
		return x.UserName
	}
	return ""
}

func (x *User) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

// Marshal 序列化
func (x *User) Marshal() ([]byte, error) {
	var buf []byte
	if x.UserName != "" {
		buf = protowire.AppendTag(buf, 1, protowire.BytesType)
		buf = protowire.AppendString(buf, x.UserName)
	}
	if x.Password != "" {
		buf = protowire.AppendTag(buf, 2, protowire.BytesType)
		buf = protowire.AppendString(buf, x.Password)
	}
	return buf, nil
}

// Unmarshal 反序列化
func (x *User) Unmarshal(data []byte) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		data = data[n:]

		switch num {
		case 1:
			if typ == protowire.BytesType {
				v, n := protowire.ConsumeString(data)
				if n < 0 {
					return protowire.ParseError(n)
				}
				x.UserName = v
				data = data[n:]
			}
		case 2:
			if typ == protowire.BytesType {
				v, n := protowire.ConsumeString(data)
				if n < 0 {
					return protowire.ParseError(n)
				}
				x.Password = v
				data = data[n:]
			}
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
		}
	}
	return nil
}
