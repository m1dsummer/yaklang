package iiop

import "github.com/yaklang/yaklang/common/yak/yaklib/codec"

const (
	CallRemoteFunS  = "47494f50010200000000012a0000000503000000000000000000008800424541080103000000000c41646d696e536572766572000000000000000044524d493a7765626c6f6769632e636c75737465722e73696e676c65746f6e2e436c75737465724d617374657252656d6f74653a30303030303030303030303030303030000000000432393300000000014245412c000000100000000000000000f4a91a4bf7a44c4a000000126765745365727665724c6f636174696f6e00000000000003000000050000001800000000000000010000000a3132372e302e302e3100e4f4000000010000000c0000000000010020050100014245410000000004000a03067fffff020000002349444c3a6f6d672e6f72672f434f5242412f57537472696e6756616c75653a312e3000000000000677686f616d69"
	LocateRequestS  = "47494f50010200030000001700000002000000000000000b4e616d6553657276696365"
	RebindS         = "47494f5001020000000005ac0000000203000000000000000000007800424541080103000000000c41646d696e53657276657200000000000000003349444c3a7765626c6f6769632f636f7262612f636f732f6e616d696e672f4e616d696e67436f6e74657874416e793a312e3000000000000238000000000000014245412a0000001000000000000000008dc24b6eacfc6f850000000b726562696e645f616e79000000000006000000050000001800000000000000010000000a3132372e302e302e3100ff8a000000010000000c00000000000100200501000100000006000000f0000000000000002849444c3a6f6d672e6f72672f53656e64696e67436f6e746578742f436f6465426173653a312e30000000000100000000000000b4000102000000000a3132372e302e302e3100ff8a0000006400424541080103000000000100000000000000000000002849444c3a6f6d672e6f72672f53656e64696e67436f6e746578742f436f6465426173653a312e30000000000331320000000000014245412a000000100000000000000000182e29cfed78fb1a00000001000000010000002c00000000000100200000000300010020000100010501000100010100000000030001010000010109050100010000000f00000020000000000000000000000000000000010000000000000000010000000000000042454103000000140000000000000000ffffffffacfc6f85000000004245410000000004000a030600000000000000010000000568656c6c6f00000000000001000000000000001d0000001c000000000000000100000000000000010000000000000000000000007fffff020000003e524d493a7765626c6f6769632e69696f702e50726f7879446573633a373343443941343543424135323933383a373432363138303142393331454630300000007fffff0200000059524d493a73756e2e7265666c6563742e616e6e6f746174696f6e2e416e6e6f746174696f6e496e766f636174696f6e48616e646c65723a433030334245443736453333333842423a35354341463530463135434237454135000000007fffff0a00000038524d493a6a6176612e7574696c2e486173684d61703a383635373335363841323131433031313a303530374441433143333136363044310000000015010100003f4000000000000c0000001000000001000000007fffff0a0000002349444c3a6f6d672e6f72672f434f5242412f57537472696e6756616c75653a312e300000000000180000001470776e6564323531343134383730303733373931fffffffe00000001000000007fffff0a00000074524d493a636f6d2e6265612e636f72652e72657061636b616765642e737072696e676672616d65776f726b2e7472616e73616374696f6e2e6a74612e4a74615472616e73616374696f6e4d616e616765723a304433303438453037423144334237423a34454633454346424236323839383246000000001cffffffff0001010000000000000001010100000000000000000000007fffff0affffffffffffff0c0000002600000022726d693a2f2f3139322e3136382e3130312e3131363a393039302f616263646561610000ffffffff7fffff0200000040524d493a6a617661782e726d692e434f5242412e436c617373446573633a324241424441303435383741444343433a43464246303243463532393431373642007fffff02fffffffffffffe88000000007fffff02fffffffffffffe7800000027524d493a6a6176612e6c616e672e4f766572726964653a30303030303030303030303030303030007fffff0200000039524d493a5b4c6a6176612e6c616e672e436c6173733b3a303731444138424537463937313132383a3243374535353033443942463935353300000000000000017fffff02ffffffffffffff24ffffffffffffff607fffff02fffffffffffffde400000024524d493a6a6176612e726d692e52656d6f74653a30303030303030303030303030303030"
	BindS           = "47494f5001020000000005ac0000000203000000000000000000007800424541080103000000000c41646d696e53657276657200000000000000003349444c3a7765626c6f6769632f636f7262612f636f732f6e616d696e672f4e616d696e67436f6e74657874416e793a312e3000000000000238000000000000014245412a000000100000000000000000d934a7566f1df5840000000962696e645f616e790000000000000006000000050000001800000000000000010000000a3132372e302e302e3100d61a000000010000000c00000000000100200501000100000006000000f0000000000000002849444c3a6f6d672e6f72672f53656e64696e67436f6e746578742f436f6465426173653a312e30000000000100000000000000b4000102000000000a3132372e302e302e3100d61a0000006400424541080103000000000100000000000000000000002849444c3a6f6d672e6f72672f53656e64696e67436f6e746578742f436f6465426173653a312e30000000000331320000000000014245412a0000001000000000000000008c8fccd18886d2d600000001000000010000002c00000000000100200000000300010020000100010501000100010100000000030001010000010109050100010000000f00000020000000000000000000000000000000010000000000000000010000000000000042454103000000140000000000000000000000006f1df584000000004245410000000004000a030600000000000000010000000568656c6c6f00000000000001000000000000001d0000001c000000000000000100000000000000010000000000000000000000007fffff020000003e524d493a7765626c6f6769632e69696f702e50726f7879446573633a373343443941343543424135323933383a373432363138303142393331454630300000007fffff0200000059524d493a73756e2e7265666c6563742e616e6e6f746174696f6e2e416e6e6f746174696f6e496e766f636174696f6e48616e646c65723a433030334245443736453333333842423a35354341463530463135434237454135000000007fffff0a00000038524d493a6a6176612e7574696c2e486173684d61703a383635373335363841323131433031313a303530374441433143333136363044310000000015010100003f4000000000000c0000001000000001000000007fffff0a0000002349444c3a6f6d672e6f72672f434f5242412f57537472696e6756616c75653a312e300000000000180000001470776e6564343630353738323836373637303431fffffffe00000001000000007fffff0a00000074524d493a636f6d2e6265612e636f72652e72657061636b616765642e737072696e676672616d65776f726b2e7472616e73616374696f6e2e6a74612e4a74615472616e73616374696f6e4d616e616765723a304433303438453037423144334237423a34454633454346424236323839383246000000001cffffffff0001010000000000000001010100000000000000000000007fffff0affffffffffffff0c0000002600000022726d693a2f2f3139322e3136382e3130312e3131363a393039302f4578706c6f69740000ffffffff7fffff0200000040524d493a6a617661782e726d692e434f5242412e436c617373446573633a324241424441303435383741444343433a43464246303243463532393431373642007fffff02fffffffffffffe88000000007fffff02fffffffffffffe7800000027524d493a6a6176612e6c616e672e4f766572726964653a30303030303030303030303030303030007fffff0200000039524d493a5b4c6a6176612e6c616e672e436c6173733b3a303731444138424537463937313132383a3243374535353033443942463935353300000000000000017fffff02ffffffffffffff24ffffffffffffff607fffff02fffffffffffffde400000024524d493a6a6176612e726d692e52656d6f74653a30303030303030303030303030303030"
	Reslove_anyS    = "47494f5001020000000002350000000203000000000000000000007800424541080103000000000c41646d696e53657276657200000000000000003349444c3a7765626c6f6769632f636f7262612f636f732f6e616d696e672f4e616d696e67436f6e74657874416e793a312e3000000000000238000000000000014245412c000000100000000000000000f4a91a4bf7a44c4a0000000c7265736f6c76655f616e790000000006000000050000001800000000000000010000000a3132372e302e302e3100c792000000010000000c00000000000100200501000100000006000000f0000000000000002849444c3a6f6d672e6f72672f53656e64696e67436f6e746578742f436f6465426173653a312e30000000000100000000000000b4000102000000000a3132372e302e302e3100c7920000006400424541080103000000000100000000000000000000002849444c3a6f6d672e6f72672f53656e64696e67436f6e746578742f436f6465426173653a312e30000000000331320000000000014245412a000000100000000000000000daa74837e81b613500000001000000010000002c00000000000100200000000300010020000100010501000100010100000000030001010000010109050100010000000f00000020000000000000000000000000000000010000000000000000010000000000000042454103000000140000000000000000fffffffff7a44c4a000000004245410000000004000a030600000000000000010000000b556e69636f646553656361000000000100"
	RebindBackdoorS = "47494f5001020000000010240000000303000000000000000000007800424541080103000000000c41646d696e53657276657200000000000000003349444c3a7765626c6f6769632f636f7262612f636f732f6e616d696e672f4e616d696e67436f6e74657874416e793a312e3000000000000432383900000000014245412c0000001000000000000000004f910dd135dbb67e0000000b726562696e645f616e79000000000004000000050000001800000000000000010000000a3132372e302e302e3100f3b6000000010000000c000000000001002005010001424541030000001400000000000000000000000035dbb67e000000004245410000000004000a030600000000000000010000001a556e69636f646553656361323631343530333235393631363235000000000001000000000000001d0000001c000000000000000100000000000000010000000000000000000000007fffff020000005a524d493a636f6d2e74616e676f736f6c2e696e7465726e616c2e7574696c2e696e766f6b652e52656d6f7465436f6e7374727563746f723a433542324538413746343945443846463a434439414543343635424445454137370000007fffff0200000029524d493a5b4c6a6176612e6c616e672e4f626a6563743b3a3030303030303030303030303030303000000000000000007fffff0200000058524d493a636f6d2e74616e676f736f6c2e696e7465726e616c2e7574696c2e696e766f6b652e436c617373446566696e6974696f6e3a413531304136393033324541454244443a32354342424538424637303044413941007fffff0200000018524d493a5b423a303030303030303030303030303030300000000cdccafebabe0000003300ab0a002600620800630a006400650a006600670800680a0066006907006a0a0007006208006b0b006c006d08006e08006f0800700700710a000e00720a000e00730a000e00740700750700760a007700780a001300790a0012007a07007b0a001700620a0012007c0a0017007d08007e0a0017007f0700800a001d00810800820700830a002000620700a90a002200620b008500860a001d008707008807008907008a0100063c696e69743e010003282956010004436f646501000f4c696e654e756d6265725461626c650100124c6f63616c5661726961626c655461626c650100047468697301000f4c636f6d2f6b616d692f746573743b01001467657452656d6f7465436f6e7374727563746f7201003728294c636f6d2f74616e676f736f6c2f696e7465726e616c2f7574696c2f696e766f6b652f52656d6f7465436f6e7374727563746f723b01001473657452656d6f7465436f6e7374727563746f72010038284c636f6d2f74616e676f736f6c2f696e7465726e616c2f7574696c2f696e766f6b652f52656d6f7465436f6e7374727563746f723b295601001172656d6f7465436f6e7374727563746f720100354c636f6d2f74616e676f736f6c2f696e7465726e616c2f7574696c2f696e766f6b652f52656d6f7465436f6e7374727563746f723b0100117365745365727665724c6f636174696f6e010027284c6a6176612f6c616e672f537472696e673b4c6a6176612f6c616e672f537472696e673b2956010004766172310100124c6a6176612f6c616e672f537472696e673b0100047661723201000a457863657074696f6e7307008b0100116765745365727665724c6f636174696f6e010026284c6a6176612f6c616e672f537472696e673b294c6a6176612f6c616e672f537472696e673b01000769734c696e75780100015a0100056f73547970010004636d64730100104c6a6176612f7574696c2f4c6973743b01000e70726f636573734275696c64657201001a4c6a6176612f6c616e672f50726f636573734275696c6465723b01000470726f630100134c6a6176612f6c616e672f50726f636573733b01000262720100184c6a6176612f696f2f42756666657265645265616465723b01000273620100184c6a6176612f6c616e672f537472696e674275666665723b0100046c696e65010001650100154c6a6176612f6c616e672f457863657074696f6e3b010003636d640100164c6f63616c5661726961626c65547970655461626c650100244c6a6176612f7574696c2f4c6973743c4c6a6176612f6c616e672f537472696e673b3e3b01000d537461636b4d61705461626c6507008c07008d0700a907007107008e07007507007b0700800100083c636c696e69743e01000862696e644e616d650100036374780100164c6a617661782f6e616d696e672f436f6e746578743b01000672656d6f746501000a536f7572636546696c65010009746573742e6a6176610c0029002a0100076f732e6e616d6507008f0c0090003e07008c0c0091009201000377696e0c009300940100136a6176612f7574696c2f41727261794c6973740100092f62696e2f6261736807008d0c009500960100022d63010007636d642e6578650100022f630100186a6176612f6c616e672f50726f636573734275696c6465720c002900970c009800990c009a009b0100166a6176612f696f2f42756666657265645265616465720100196a6176612f696f2f496e70757453747265616d52656164657207008e0c009c009d0c0029009e0c0029009f0100166a6176612f6c616e672f537472696e674275666665720c00a000920c00a100a20100010a0c00a300920100136a6176612f6c616e672f457863657074696f6e0c00a4009201000b556e69636f64655365636101001b6a617661782f6e616d696e672f496e697469616c436f6e7465787401000d636f6d2f6b616d692f746573740700a50c00a600a70c00a8002a0100106a6176612f6c616e672f4f626a65637401002b636f6d2f74616e676f736f6c2f696e7465726e616c2f7574696c2f696e766f6b652f52656d6f7461626c6501002e7765626c6f6769632f636c75737465722f73696e676c65746f6e2f436c75737465724d617374657252656d6f74650100186a6176612f726d692f52656d6f7465457863657074696f6e0100106a6176612f6c616e672f537472696e6701000e6a6176612f7574696c2f4c6973740100116a6176612f6c616e672f50726f636573730100106a6176612f6c616e672f53797374656d01000b67657450726f706572747901000b746f4c6f7765724361736501001428294c6a6176612f6c616e672f537472696e673b010008636f6e7461696e7301001b284c6a6176612f6c616e672f4368617253657175656e63653b295a010003616464010015284c6a6176612f6c616e672f4f626a6563743b295a010013284c6a6176612f7574696c2f4c6973743b295601001372656469726563744572726f7253747265616d01001d285a294c6a6176612f6c616e672f50726f636573734275696c6465723b010005737461727401001528294c6a6176612f6c616e672f50726f636573733b01000e676574496e70757453747265616d01001728294c6a6176612f696f2f496e70757453747265616d3b010018284c6a6176612f696f2f496e70757453747265616d3b2956010013284c6a6176612f696f2f5265616465723b2956010008726561644c696e65010006617070656e6401002c284c6a6176612f6c616e672f537472696e673b294c6a6176612f6c616e672f537472696e674275666665723b010008746f537472696e6701000a6765744d6573736167650100146a617661782f6e616d696e672f436f6e74657874010006726562696e64010027284c6a6176612f6c616e672f537472696e673b4c6a6176612f6c616e672f4f626a6563743b295601000f7072696e74537461636b547261636501002e636f6d2f6b616d692f746573742441344342433242443037383933464139464539414535373144324132304533460100304c636f6d2f6b616d692f746573742441344342433242443037383933464139464539414535373144324132304533463b0021002200260002002700280000000600010029002a0001002b0000003300010001000000052ab70001b100000002002c0000000a00020000001c0004001e002d0000000c000100000005002e00aa00000001003000310001002b0000002c000100010000000201b000000002002c00000006000100000022002d0000000c000100000002002e00aa00000001003200330001002b000000350000000200000001b100000002002c00000006000100000028002d00000016000200000001002e00aa0000000000010034003500010001003600370002002b0000003f0000000300000001b100000002002c0000000600010000002d002d00000020000300000001002e00aa00000000000100380039000100000001003a00390002003b000000040001003c0001003d003e0002002b0000020a0005000a000000c2043d1202b800034e2dc600112db600041205b60006990005033dbb000759b700083a041c99002319041209b9000a0200571904120bb9000a02005719042bb9000a020057a700201904120cb9000a0200571904120db9000a02005719042bb9000a020057bb000e591904b7000f3a05190504b60010571905b600113a06bb001259bb0013591906b60014b70015b700163a07bb001759b700183a081907b60019593a09c6001319081909b6001a121bb6001a57a7ffe81908b6001cb04d2cb6001eb00001000000bb00bc001d0004002c0000005a001600000033000200340008003500180036001a00380023003a0027003b0031003c003b003d0047003f00510040005b004100640044006f004500760046007d004800920049009b004c00a6004d00b6005000bc005100bd0052002d00000070000b000200ba003f00400002000800b400410039000300230099004200430004006f004d004400450005007d003f0046004700060092002a004800490007009b0021004a004b000800a30019004c0039000900bd0005004d004e0002000000c2002e00aa0000000000c2004f0039000100500000000c0001002300990042005100040052000000460006fd001a01070053fc002c0700541cff00360009070055070053010700530700540700560700570700580700590000fc001a070053ff00050002070055070053000107005a003b000000040001003c0008005b002a0001002b0000009d0003000300000024121f4bbb002059b700214cbb002259b700234d2b2a2cb900240300a700084b2ab60025b100010000001b001e001d0003002c0000002200080000001300030014000b001500130016001b0019001e0017001f00180023001a002d0000002a000400030018005c00390000000b0010005d005e000100130008005f00aa0002001f00040038004e000000520000000700025e07005a04000100600000000200617fffff0200000056524d493a636f6d2e74616e676f736f6c2e696e7465726e616c2e7574696c2e696e766f6b652e436c6173734964656e746974793a313132374635323437464144443241463a374434364331373443304634323833390000007fffff020000002349444c3a6f6d672e6f72672f434f5242412f57537472696e6756616c75653a312e30000000000004746573747fffff02ffffffffffffffc800000008636f6d2f6b616d697fffff02ffffffffffffffb0000000204134434243324244303738393346413946453941453537314432413230453346"
)

func GetCallRemoteFunMsgTmp() *MessageRequest {
	CallRemoteFun, _ := codec.DecodeHex(CallRemoteFunS)
	CallRemoteFunMsg, _ := ParseMessageRequest(CallRemoteFun)
	return CallRemoteFunMsg
}
func GetLocateRequestMsgTmp() *MessageRequest {
	LocateRequest, _ := codec.DecodeHex(LocateRequestS)
	LocateRequestMsg, _ := ParseMessageRequest(LocateRequest)
	return LocateRequestMsg
}

func GetRebindMsgTmp() *MessageRequest {
	Rebind, _ := codec.DecodeHex(RebindS)
	RebindMsg, _ := ParseMessageRequest(Rebind)
	return RebindMsg
}
func GetReslove_anyMsgTmp() *MessageRequest {
	Reslove_any, _ := codec.DecodeHex(Reslove_anyS)
	Reslove_anyMsg, _ := ParseMessageRequest(Reslove_any)
	return Reslove_anyMsg
}
func GetBindMsgTmp() *MessageRequest {
	Bind, _ := codec.DecodeHex(BindS)
	BindMsg, _ := ParseMessageRequest(Bind)
	return BindMsg
}
func GetRebindBackdoorMsgTmp() *MessageRequest {
	RebindBackdoor, _ := codec.DecodeHex(RebindBackdoorS)
	RebindBackdoorMsg, _ := ParseMessageRequest(RebindBackdoor)
	return RebindBackdoorMsg
}