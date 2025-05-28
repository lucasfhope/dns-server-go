# DNS Server

# Go HTTP Server

This is my implementation of a simple DNS server, built in Go, by following the [Codecrafters "Build Your Own DNS Server" challenge](https://app.codecrafters.io/courses/dns-server). The DNS server utilzes UDP connection for fast queries and responses that do not require a handshake like TCP does. My DNS server can take in messages and send reponses, and it also can connect and a send queries through to another DNS server.

---

## Table of Contents

1. [Getting Started](#getting-started)
   - [Requirements](#requirements)
   - [Quickstart](#quickstart)
2. [Usage](#usage)


---

# Getting Started

## Requirements

- **git**
    - Try running `git --version` to see if it is installed
- **go**
    - Try running `go version` to se if it is installed

## Quickstart

```bash
git clone https://github.com/lucasfhope/go-dns-server.git
cd go-dns-server
```

---

# Usage

The DNS server lists to the loopback IP address `127.0.0.1` on port `2053`.

## Start the server

Run `app/main.go` to start the server.

```bash
go run app/main.go
```

## Interact with the server

In another terminal, you can interact with the server with the IP address `127.0.0.1` on port `2053`.

My DNS server expects a proper DNS message with a header and a question section. All values in the DNS messages are encoded as big-endian integers.

A DNS header is 12 bytes long. It contains the following fields:

| Field |	Size | Expected Value |
| ----- | ---- | --------------- |
| Packet Identifier (ID) | 16 bits | Any 16 bit identifier |
| Query/Response Indicator (QR) |	1 bit	| 0 for query, 1 for response |
| Operation Code (OPCODE) | 4 bits | My server has only implemented 0000 |
| Authoritative Answer (AA)	| 1 bit	| Leave as 0 |
| Truncation (TC)	| 1 bit	| Leave as 0 |
| Recursion Desired (RD) | 1 bit	| 0 or 1 |
| Recursion Available (RA)	| 1 bit	| Leave a 0 |
| Reserved (Z)	| 3 bits | Leave as 000 |
| Response Code (RCODE)	| 4 bits | Leave as 0000 |
| Question Count (QDCOUNT)	| 16 bits	| Number of questions in the message |
| Answer Record Count (ANCOUNT) |	16 bits	| Number of answers in the message |
| Authority Record Count (NSCOUNT) | 16 bits	| Leave as 0000 |
| Additional Record Count (ARCOUNT)	| 16 bits	| Leave as 0000 |

Each DNS question should contain a domain name as a label sequence, a type (2-byte integer), and a class (2-byte integer). The domain name should broken up into labels by seperating each label at the `.`. Each label should be prepended with 1 byte for the length of the label, and then contain the hexadecimal byte value for each character. At the end of the label sequence, append a null byte (0x00) to mark the end of the domain name.

To send a DNS query for `example.com`, you could enter the command:

```bash
echo -n '04d201000001000000000000076578616D706C6503636F6D0000010001' | xxd -r -p | nc -u -w 1 127.0.0.1 2053 | hexdump -C
```

## Server Response

My DNS server will respond differently based on the query message it receive. For the header, the server will mimic the packet identifier, have a `1` for the response indicator, mimic the OPCODE, and mimic the recursion desired field. The response code will be `0000` if the the OPCODE is `0000`, and it will be `0004` otherwise to indicate that the OPCODE functionality has not been implemented. Question count and answer count will be based on the number of questions in the DNS query. All other fields will be 0.

The DNS response will also contain a question and answer that corresponds to each of the questions in the DNS query.

Each question in the response will be the same as the question in the query, with the domain name as a label sequence. Note that my DNS server can only handle type `0001` ('A') and class `0001` ('IN') and will always respond with these values for these fields, even if the query message specifies other types or classes. 

The answer in the response will also contain the domain name as a label sequence, type `0001`, and class `0001`. The next field is the time-to-live (TTL), which is a 4-byte integer, indiacting how long (in seconds) the IP address can be associated with a domain name before a refresh. My server will always respond with `67543`. The next field is the length of the data field encoded as a 2-byte integer, which is always `4` since my server only handles IPv4 requests. The final field in the answer is the data section, encoded as a 4-byte integer, which is the IPv4 address. My server will always respond with `0.0.0.0`, or `\0x00\0x00\0x00\0x00`.

If you sent the DNS message above, `04d201000001000000000000076578616D706C6503636F6D0000010001`, you should recieve this response:

```
04 d2 81 00 00 01 00 01  00 00 00 00 07 65 78 61
6d 70 6c 65 03 63 6f 6d  00 00 01 00 01 c0 0c 00 
01 00 01 00 01 07 d7 00  04 00 00 00 00 
```

---

# Other Features

## Compression

My DNS server can handle messages with compressed label sequences and can responsed with a compressed DNS packet. 

Compression works in the domain name label sequences. If a label or sequance has been encoded before and needs to be encoded again, two bytes can be used as a pointer to the beginning of where this label was previously encoded. If you look at the byte that usually dentoes the length of the label, and its first two bits `11`, then that byte and the next are actually a pointer to a different location in the DNS packet, where 0 is pointing to the first byte of the packet identifier.

This is why, in the example above, the domain name in the DNS response answer section was `c0 0c` or `1100 0011`. The `11` in the beginning indicated a pointer, and the remaining bits, `00 0011` was equal to 12, pointing to the start of the question section. This is where the label sequance for `example.com` was located. 

If you want to send a DNS packet that contains a pointer, consider sending questions for both `example.com` and `it.example.com`. First, indiacate that you are sending two questions by updating the QDCount field to `0002`. Then, for the second question, encode `it` as `\0x02\0x69\0x74`. Then create a pointer to DNS packet position 12 with `\0xC0\0x0C`. Finish the second question with the type and class fields.

```bash
echo -n '04d201000002000000000000076578616D706C6503636F6D0000010001026974C00C00010001' | xxd -r -p | nc -u -w 1 127.0.0.1 2053 | hexdump -C
```

You should recieve the following response with two pointers to position 12, and one pointer to position 29 (`c0 1d`), which is the beginning of the second question section. The full response is below.

```
04 d2 81 00 00 02 00 02  00 00 00 00 07 65 78 61
6d 70 6c 65 03 63 6f 6d  00 00 01 00 01 02 69 74
c0 0c 00 01 00 01 c0 0c  00 01 00 01 00 01 07 d7
00 04 00 00 00 00 c0 1d  00 01 00 01 00 01 07 d7
00 04 00 00 00 00
```

## Server Forwarding

To forward the DNS queries to another DNS server, use the `--resolver` flag.

```bash
go run main.go --resolver <address>
```

To forward to Google's DNS server, use the address `1.1.1.1` on port `53`.

```bash
go run main.go --resolver 1.1.1.1:53
```
