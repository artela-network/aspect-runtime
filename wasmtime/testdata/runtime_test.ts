// The entry file of your WebAssembly module.

export declare namespace test {
    function hello(ptr: i32): i32
    function hello2(ptr1: i32, ptr2: i32, ptr3: i32): i32
    function log(s: string): void
    function hello3(ptr: i32): void
}

enum typeIndex {
    Empty = 0,
    TypeInt8,
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

class AUint8Array {
    head: header;
    body: Uint8Array; // utf-16 encoder
    constructor(head: header = new header(typeIndex.TypeByteArray, 0), body: Uint8Array = new Uint8Array(0)) {
        this.head = head;
        this.body = body;
    }

    set(data: Uint8Array): void {
        this.body = data;
        this.head.dataLen = data.length;
    }

    get(): Uint8Array {
        return this.body;
    }

    load(ptr: i32): void {
        this.head = new header(0, 0);
        this.head.load(ptr);
        let bodyPtr = ptr + this.head.len();
        this.body = new Uint8Array(this.head.dataLen);
        for (let i = 0; i < this.head.dataLen; i++) {
            this.body[i] = u8(i32.load8_u(bodyPtr));
            bodyPtr++;
        }
    }

    store(): i32 {
        let ptr = allocate(this.head.dataLen + this.head.len())
        this.head.store(ptr);
        let bodyPtr = ptr + this.head.len();
        for (let i = 0; i < this.head.dataLen; i++) {
            memory.fill(bodyPtr, this.body[i], 1)
            bodyPtr++;
        }
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

export function greet2(ptr1: i32, ptr2: i32, ptr3: i32): i32 {
    let req1 = new AString();
    req1.load(ptr1);
    let req2 = new AString();
    req2.load(ptr2);
    let req3 = new AString();
    req3.load(ptr3);

    let helloReq1 = new AString();
    helloReq1.set(req1.get())
    let helloArg1 = helloReq1.store()
    let helloReq2 = new AString();
    helloReq2.set(req2.get())
    let helloArg2 = helloReq2.store()
    let helloReq3 = new AString();
    helloReq3.set(req3.get())
    let helloArg3 = helloReq3.store()

    let helloRetPtr = test.hello2(helloArg1, helloArg2, helloArg3)
    let helloRet = new AString();
    helloRet.load(helloRetPtr);

    let retAs = new AString();
    retAs.set(helloRet.get() + "-over");
    let retPtr = retAs.store();
    return retPtr;
}

export function greet3(ptr: i32): i32 {
    let req = new AString();
    req.load(ptr);
    let helloReq1 = new AString();
    helloReq1.set("greet3-" + req.get());
    let helloptr = helloReq1.store();
    test.hello3(helloptr);
    let out = new AString();
    out.set("greet3");
    return out.store();
}

export function testBytes(ptr: i32): i32 {
    let req = new AUint8Array();
    req.load(ptr);
    for (let i = 0; i < req.head.dataLen; i++) {
        req.body[i] += 1;
    }

    let res = new AUint8Array();
    res.set(req.body);
    return res.store();
}

let sum = 0;
export function testIncrease(): i32 {
    sum = sum + 10;
    let retAs = new AString();
    retAs.set(sum.toString());
    return retAs.store();
}

export function allocate(size: i32): i32 {
    return heap.alloc(size) as i32;
}
