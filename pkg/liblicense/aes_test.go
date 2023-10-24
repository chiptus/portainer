package liblicense

import (
	"testing"
)

func TestAESEncrypt(t *testing.T) {
	input := "qwertyuiop12345678DFKJASDKJASDFKJHASDKJFHNWEM90asdfghjk"

	expected := []byte{180, 232, 74, 83, 82, 133, 176, 251, 174, 213, 191, 182, 237, 20, 81, 144, 6, 121, 50, 158, 202, 191, 96, 113, 117, 32, 12, 142, 193, 72, 168, 52, 164, 171, 87, 232, 96, 55, 203, 198, 231, 98, 128, 1, 164, 109, 78, 70, 5, 87, 183, 231, 134, 208, 85, 51, 215, 118, 10, 129, 98, 103, 240, 39, 45, 75, 103, 79, 65, 236, 163, 204, 129, 111, 150, 203, 3, 193, 238, 101, 185, 120, 64}
	got, _ := AESEncrypt([]byte(input), []byte(MasterKeyV2))

	if string(got) != string(expected) {
		t.Errorf("Expected %v but got %v", expected, got)
	}
}
