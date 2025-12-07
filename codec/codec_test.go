package codec

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/jxncyjq/stardust.mini/codec/fbs"
	"github.com/jxncyjq/stardust.mini/codec/proto"
)

func TestProtobufCodec(t *testing.T) {
	// 创建User
	user := &proto.User{
		UserName: "test_user",
		Password: "test_password",
	}

	// 序列化
	data, err := user.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	t.Logf("Protobuf encoded size: %d bytes", len(data))

	// 反序列化
	decoded := &proto.User{}
	err = decoded.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 验证
	if decoded.GetUserName() != user.GetUserName() {
		t.Errorf("UserName mismatch: got %s, want %s", decoded.GetUserName(), user.GetUserName())
	}
	if decoded.GetPassword() != user.GetPassword() {
		t.Errorf("Password mismatch: got %s, want %s", decoded.GetPassword(), user.GetPassword())
	}
	t.Logf("Protobuf test passed: UserName=%s, Password=%s", decoded.GetUserName(), decoded.GetPassword())
}

func TestFlatbufferCodec(t *testing.T) {
	codec := NewFlatbufferCodec()

	// 序列化
	builder := codec.NewBuilder(256)
	userName := builder.CreateString("test_user")
	password := builder.CreateString("test_password")

	fbs.UserStart(builder)
	fbs.UserAddUserName(builder, userName)
	fbs.UserAddPassword(builder, password)
	userOffset := fbs.UserEnd(builder)
	builder.Finish(userOffset)

	data := codec.EncodeFlatbuffer(builder)
	t.Logf("Flatbuffer encoded size: %d bytes", len(data))

	// 反序列化
	decoded := fbs.GetRootAsUser(data, 0)

	// 验证
	if string(decoded.UserName()) != "test_user" {
		t.Errorf("UserName mismatch: got %s, want test_user", decoded.UserName())
	}
	if string(decoded.Password()) != "test_password" {
		t.Errorf("Password mismatch: got %s, want test_password", decoded.Password())
	}
	t.Logf("Flatbuffer test passed: UserName=%s, Password=%s", decoded.UserName(), decoded.Password())
}

func BenchmarkProtobufEncode(b *testing.B) {
	user := &proto.User{UserName: "test_user", Password: "test_password"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.Marshal()
	}
}

func BenchmarkFlatbufferEncode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := flatbuffers.NewBuilder(256)
		userName := builder.CreateString("test_user")
		password := builder.CreateString("test_password")
		fbs.UserStart(builder)
		fbs.UserAddUserName(builder, userName)
		fbs.UserAddPassword(builder, password)
		userOffset := fbs.UserEnd(builder)
		builder.Finish(userOffset)
	}
}
