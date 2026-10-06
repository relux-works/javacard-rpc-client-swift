import Foundation
import Testing
@testable import JavaCardRPCClient

// Runtime framing and typed readers preserve the existing short APDU wire contract.
@Test func commandAndResponseWireFormat() throws {
    var packer = DataPacker()
    packer.packU8(0xAB)
    packer.packU16(0x1234)
    packer.packU32(0x56789ABC)
    packer.packBool(false)
    packer.packBool(true)
    packer.packBytes(Data([0xEE]))
    #expect(packer.data == Data([0xAB, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0, 1, 0xEE]))
    #expect(APDUCommand(cla: 0xB0, ins: 1, p1: 2, p2: 3, data: packer.data, le: 0).bytes == Data([0xB0, 1, 2, 3, 10]) + packer.data + Data([0]))
    #expect(APDUCommand(cla: 0, ins: 0xA4, data: Data()).bytes == Data([0, 0xA4, 0, 0]))
    let response = APDUResponse(rawBytes: packer.data + Data([0x90, 0]))
    try response.checkSW()
    #expect(response.readU8() == 0xAB)
    #expect(response.readU16(at: 1) == 0x1234)
    #expect(response.readU32(at: 3) == 0x56789ABC)
    #expect(!response.readBool(at: 7))
    #expect(response.readBool(at: 8))
    #expect(response.readBytes(at: 9, count: 1) == Data([0xEE]))
    #expect(APDUResponse(rawBytes: Data([0x90, 0])).data.isEmpty)
}

// Non-success status words retain their exact error and must never be admitted as success.
@Test(arguments: [UInt16(0x6985), 0x6A80, 0x9001])
func statusWordRejection(word: UInt16) throws {
    let response = APDUResponse(rawBytes: Data([UInt8(word >> 8), UInt8(word & 0xFF)]))
    do {
        try response.checkSW()
        Issue.record("Non-success status word admitted")
    } catch APDUError.statusWord(let sw1, let sw2) {
        #expect(sw1 == UInt8(word >> 8))
        #expect(sw2 == UInt8(word & 0xFF))
    }
    try APDUResponse(rawBytes: Data([0x90, 0])).checkSW()
}

// A wrong fixed-byte count rejects before appending; nearby valid counts still append.
@Test(arguments: [1, 3])
func fixedBytesRejectWithoutPartialWrite(count: Int) throws {
    var packer = DataPacker()
    packer.packU8(0xAB)
    do {
        try packer.packFixedBytes(Data(repeating: 0xCC, count: count), length: 2)
        Issue.record("Wrong fixed-byte length admitted")
    } catch APDUError.invalidResponse {
        #expect(packer.data == Data([0xAB]))
    }
    try packer.packFixedBytes(Data([1, 2]), length: 2)
    #expect(packer.data == Data([0xAB, 1, 2]))
}
