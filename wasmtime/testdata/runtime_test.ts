// The entry file of your WebAssembly module.

export declare namespace test {
    function hello(ptr1: i32, ptr2: i32): i32
    function log(s: string): void
}

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

export function greet(ptr: i32): string {
    let input = helper.unpackMsg(ptr);
    test.log(input);
    // TODO packmessage
    let hinput = allocate(10);
    test.log(hinput.toString());
    memory.fill(hinput, 10, 1);
    memory.fill(hinput + 1, 0, 1);
    memory.fill(hinput + 2, 4, 1)
    memory.fill(hinput + 3, 0, 3);

    memory.fill(hinput + 6, 65, 1);
    memory.fill(hinput + 7, 66, 1);
    memory.fill(hinput + 8, 67, 1);
    memory.fill(hinput + 9, 68, 1);

    let hinput2 = allocate(10);
    test.log(hinput2.toString());
    memory.fill(hinput2, 9, 1);
    memory.fill(hinput2 + 1, 0, 1);
    memory.fill(hinput2 + 2, 4, 1)
    memory.fill(hinput2 + 3, 0, 3);

    memory.fill(hinput2 + 6, 97, 1);
    memory.fill(hinput2 + 7, 98, 1);
    memory.fill(hinput2 + 8, 99, 1);
    memory.fill(hinput2 + 9, 100, 1);
    let p = test.hello(hinput, hinput2);
    // test.log(p.toString());
    let s = helper.unpackMsg(ptr);
    test.log(s);
    return "hello" + s;
}


export function allocate(size: i32): i32 {
    return heap.alloc(size) as i32;
}