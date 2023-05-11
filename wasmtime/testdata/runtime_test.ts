// The entry file of your WebAssembly module.

export declare namespace test {
    function hello(ptr: i32): i32
    function log(s: string): void
}

enum typeIndex {
    TypeInt8 = 0,
    TypeInt16,
    TypeInt32,
    TypeInt64,
    TypeUint8,
    TypeUint16,
    TypeUint32,
    TypeUint64,
    TypeBool,
    TypeString, // string with utf-8 encoder
    TypeByteArray,
}

class header {
    dataType: i16;
    dataLen: i32;

    constructor(type: typeIndex = typeIndex.TypeInt8, len: i32 = 0) {
        this.dataType = type as i16;
        this.dataLen = len;
    }

    load(ptr: i32): void {
        this.dataType = i16(i32.load16_s(ptr));
        this.dataLen = i32.load(ptr + 2);
    }

    store(ptr: i32): void {
        i32.store16(ptr, this.dataType)
        i32.store(ptr + 2, this.dataLen)
    }

    len(): i32 {
        return 6; // i16 + i32
    }
}

class AString {
    head: header;
    body: string; // utf-16 encoder
    constructor(head: header = new header(typeIndex.TypeString, 0), body: string = "") {
        this.head = head;
        this.body = body;
    }

    set(s: string): void {
        this.body = s;
        this.head.dataLen = s.length;
    }

    get(): string {
        return this.body;
    }

    load(ptr: i32): void {
        this.head = new header(0, 0);
        this.head.load(ptr);
        let bodyPtr = ptr + this.head.len();
        // read as a utf8 string
        this.body = String.UTF8.decodeUnsafe(bodyPtr, this.head.dataLen);
    }

    store(): i32 {
        let ptr = allocate(this.head.dataLen + this.head.len())
        this.head.store(ptr);
        let bodyPtr = ptr + this.head.len();
        // utf-16 <--> utf8
        String.UTF8.encodeUnsafe(changetype<usize>(this.body), this.head.dataLen, bodyPtr)
        // it's weird that it doesn't work in the following way:
        // | let encoded = String.UTF8.encode(this.body);
        // | store<ArrayBuffer>(bodyPtr, encoded);
        return ptr;
    }
}

export function greet(ptr: i32): i32 {
    let req = new AString();
    req.load(ptr);

    let helloReq = new AString();
    helloReq.set("greet-" + req.get())
    let helloArg = helloReq.store()
    let helloRetPtr = test.hello(helloArg);
    let helloRet = new AString();
    helloRet.load(helloRetPtr);

    let retAs = new AString();
    retAs.set(helloRet.get() + "-greet");
    let retPtr = retAs.store();
    return retPtr;
}


export function allocate(size: i32): i32 {
    return heap.alloc(size) as i32;
}