# DNS Server

## Introduction
A **DNS (Domain Name System) server** is a critical component of the internet that translates human-readable domain names (e.g., `example.com`) into machine-readable IP addresses (e.g., `192.168.1.1`). This project implements a basic DNS server in Go, capable of handling DNS queries for `A` (IPv4) records, forwarding queries to an upstream resolver, and responding with appropriate DNS responses.

---

## Features
- **Local DNS Resolution**: Responds to queries locally without forwarding them.
- **Resolver Forwarding**: Forwards queries to an upstream DNS resolver if specified.
- **Error Handling**: Gracefully handles invalid queries, unreachable resolvers, and unimplemented opcodes.
- **Test Coverage**: Includes unit tests to validate server behavior for various scenarios.

---

## Usage

### Running the Server
To start the DNS server, run the following command:

```bash
go run app/main.go --resolver <resolver_address>
```

- **`<resolver_address>`**: The address of an upstream DNS resolver in the format `<ip>:<port>`. Examples:
  - `8.8.8.8:53` (Google's public DNS server)
  - `1.1.1.1:53` (Cloudflare's public DNS server)

If no resolver is specified, the server will handle queries locally.

---

## Expected Outputs for Inputs

### 1. **Basic Query**
**Input**: A DNS query for `example.com` (Type `A`).

**Expected Output**:
- The server responds with:
  - **Transaction ID**: Matches the query's ID.
  - **Flags**: `QR=1` (response), `RD=1` (recursion desired), `RA=1` (recursion available).
  - **Question Section**: Echoes the query's question.
  - **Answer Section**:
    - `ANAME`: `example.com`
    - `ATYPE`: `1` (A record)
    - `RDATA`: `0.0.0.0` (or the resolved IP address if a resolver is used)
    - `TTL`: `67543`

---

### 2. **Query with Unimplemented Opcode**
**Input**: A DNS query with an unsupported opcode.

**Expected Output**:
- The server responds with:
  - **Transaction ID**: Matches the query's ID.
  - **Flags**: `QR=1` (response), `RCODE=4` (Not Implemented).
  - **Answer Section**: Empty.

---

### 3. **Query with Compressed QNAME**
**Input**: A DNS query with a compressed QNAME (e.g., using pointers).

**Expected Output**:
- The server correctly parses and responds to the query.
- The response includes:
  - **Question Section**: All questions decoded correctly.
  - **Answer Section**: Includes answers for all questions.

---
### 4. **Query with Multiple Questions**
**Input**: A DNS query with multiple questions.

**Expected Output**:
- The server splits the query into individual queries, forwards them to the resolver (if specified), and merges the responses.
- The response includes:
  - **Answer Section**: Answers for all questions.

---

### 5. **Invalid Resolver**
**Input**: Start the server with an invalid resolver address (e.g., `1.2.3.4:53`).

**Expected Output**:
- The server logs:
  ```
  [Failed to connect to resolver at 1.2.3.4:53]
  ```
- The server does not start.

---

### 6. **No Resolver Specified**
**Input**: Start the server without specifying a resolver.

**Expected Output**:
- The server logs:
  ```
  [DNS server listening on 127.0.0.1:2053]
  ```
- The server handles queries locally.

---


## Testing

### Running Tests
To run the test suite, use the following command:

```bash
go test ./test/...
```

### Test Scenarios
1. **Basic Query**:
   - Validates that the server responds correctly to a standard DNS query.
2. **Unimplemented Opcode**:
   - Ensures the server responds with `RCODE=4` for unsupported opcodes.
3. **Compressed QNAME**:
   - Tests the server's ability to parse and respond to queries with compressed QNAMEs.
4. **Correct Pointers**:
   - Validates that the server uses correct pointers in the response for compressed names.

---

## Resolver Address

The `--resolver` flag specifies the upstream DNS resolver to which the server forwards queries. If no resolver is provided, the server handles queries locally.

### Examples:
- **Google Public DNS**:
  ```bash
  go run app/main.go --resolver 8.8.8.8:53
  ```
- **Cloudflare DNS**:
  ```bash
  go run app/main.go --resolver 1.1.1.1:53
  ```
- **No Resolver**:
  ```bash
  go run app/main.go
  ```

---


## Example Queries

### Using `dig`
You can test the server using the `dig` command:

1. **Basic Query**:
   ```bash
   dig @127.0.0.1 -p 2053 example.com
   ```

2. **Query with Unimplemented Opcode**:
   ```bash
   dig @127.0.0.1 -p 2053 example.com +opcode=15
   ```

3. **Query with Multiple Questions**:
   ```bash
   dig @127.0.0.1 -p 2053 example.com example.org
   ```

---

## Logs and Debugging
The server logs key events, including:
- Incoming queries.
- Forwarding queries to the resolver.
- Errors (e.g., invalid resolver, parsing failures).

Example log:
```
[DNS server listening on 127.0.0.1:2053]
Received 32 bytes from 127.0.0.1
Parsed DNS request from 127.0.0.1 for example.com
Forwarding query to resolver: 8.8.8.8:53
Sent response to 127.0.0.1
```

---

## License
This project is licensed under the MIT License. See the `LICENSE` file for details.
