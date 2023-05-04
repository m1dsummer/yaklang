package tlsutils

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"testing"
)

func TestClientHello(t *testing.T) {
	rawStr := `16030100ee010000ea0303d90db6dcb831e6f519e6ea71fe907ef9d7a69d898ae4d0f64d4950d0ee236afc205ef62c1dcab075f44b08d7d1e0d27c715a3fe104ee04d3412afee3c7f1feb26e0026c02bc02fc02cc030cca9cca8c009c013c00ac014009c009d002f0035c012000a1301130213030100007b000500050100000000000a000a0008001d001700180019000b00020100000d001a0018080404030807080508060401050106010503060302010203ff0100010000120000002b0009080304030303020301003300260024001d0020fcce9d7cf031acfb4c5dc590da7d001b3734d1dda76d923a75b390de1d087201`
	raw, _ := codec.DecodeHex(rawStr)
	data, err := ParseClientHello(raw)
	if err != nil {
		panic(err)
	}
	if data.SNI() != "" {
		spew.Dump(data)
		panic("EMPTY SNI PANIC")
	}
}

func TestClientHello2(t *testing.T) {
	rawStr := `16030101f6010001f2030348125381d97807a47c6904b372c7a5299d0af761af513e8f26a84bcbddaab12220a860c121b41e59963500b04b0b56153f5a5a20ae91a2e442cf89b93a4fc2d04c003e130213031301c02cc030009fcca9cca8ccaac02bc02f009ec024c028006bc023c0270067c00ac0140039c009c0130033009d009c003d003c0035002f00ff0100016b00000014001200000f7777772e79616b6c616e672e636f6d000b000403000102000a000c000a001d0017001e00190018002300c0863e514bba3e3be7a1721e6e24f223fecdb65a7c81c8f1da6890bcaa0784293796e31accdddb89718dc05f0d86a81c95ebc713024fa71a762b93d9b75a8f54fc29f1a292637fb3a4b8c00b2c51f5c7cecae08fd250ddbcfe93db751e11152015e3c269482030bb71e53e847aa65a46b9c13aab63e38d676e324aa47d79e0378720cd87e9a7fe349921cda037c2f10a3051d0c478cb19e94663167e650b92a2a479d6e33f7c735a63964056e47b66b5f1d6195072cdcb1935457f8976b756a0230016000000170000000d0030002e040305030603080708080809080a080b080408050806040105010601030302030301020103020202040205020602002b000706030403030302002d00020101003300260024001d0020183a409f1c7d85c1ba764861fee43fe08aa687af361b6c415b5eb796b169d351`
	raw, _ := codec.DecodeHex(rawStr)
	data, err := ParseClientHello(raw)
	if err != nil {
		panic(err)
	}

	if ret := data.SNI(); ret != "www.yaklang.com" {
		spew.Dump(data)
		spew.Dump(ret)
		panic("SNI PANIC: " + ret)
	}
}

func TestHandshakeClientHello_ALPN(t *testing.T) {
	rawStr := `1603010200010001fc0303748e8ebddd187366df5da761d72d9fa0cea0666af2334ac3e90cb5e5a900a972206d0065e235e03e9a9a41d5d23ac2252a0c63fde7e63e869aed1c7d82ebeeb53d00200a0a130113021303c02bc02fc02cc030cca9cca8c013c014009c009d002f003501000193baba0000000b00020100ff010001004469000500030268320033002b0029fafa000100001d00204c4d918178d822fd7fe457c16f7744d19e1ba40310de8cd6b7f1ed2ac1b1844800170000002d00020101000a000a0008fafa001d00170018002b000706fafa03040303000d0012001004030804040105030805050108060601002300000010000e000c02683208687474702f312e31000500050100000000001b00030200020012000000000028002600002379616b6c616e672e6f73732d636e2d6265696a696e672e616c6979756e63732e636f6d9a9a000100001500b4000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000`
	raw, _ := codec.DecodeHex(rawStr)
	data, err := ParseClientHello(raw)
	if err != nil {
		panic(err)
	}
	spew.Dump(data.ALPN())
}