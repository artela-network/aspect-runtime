// The entry file of your WebAssembly module.

const HEADER_LEN = 8
const HEADER_RADIX = 16
namespace helper {
    // a lib function to read parameters from memory
    export function getString(ptr: i64, len: i64): string {
        let s = String.UTF16.decodeUnsafe(<usize>ptr, <usize>len);
        return s;
    }

    // a lib function to pack messages
    export function packMsg(raw: string): string {
        let msg = raw.length.toString(HEADER_RADIX).padStart(HEADER_LEN, "0") + raw;
        return msg;
    }

    // a lib function to unpack messags
    export function unpackMsg(p: i32): string {
        let l = u32.parse(getString(p, HEADER_LEN * 2), HEADER_RADIX);
        let raw = getString(p + HEADER_LEN * 2, l * 2);
        return raw;
    }
}

export function allocate(size: u32): u32 {
    return __alloc(size) as u32;
}

export declare namespace test {
    function hello(s: string): i32
}

// a lib function for user to call in aspect "OnTxReceived"
function Hello(s: string): string {
    // pack api arguments
    let packedMsg = s.length.toString(HEADER_RADIX).padStart(HEADER_LEN, "0") + s;

    // call api
    let ret = test.hello(packedMsg);
    let raw = helper.unpackMsg(ret);
    return raw;
}

export function test(ptr: i32): string {
    let s = helper.unpackMsg(ptr);
    let out = Hello(s);
    return helper.packMsg(out);
}