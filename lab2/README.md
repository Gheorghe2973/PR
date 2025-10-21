# Multithreaded HTTP Server Lab Report

## Student Information

* **Name:** George
* **Lab:** Multithreaded HTTP Server with Concurrency Control
* **Date:** October 17, 2025

---

## Overview

This lab extends the HTTP file server from Lab 1 by implementing multithreading, demonstrating race conditions, and adding thread-safe request counting and rate limiting features. The server demonstrates the difference between single-threaded and multithreaded architectures, as well as proper synchronization mechanisms.

---

## 1. Understanding the Four Server Implementations

This lab includes four different server implementations, each serving a specific pedagogical purpose:

### 1.1 Feature Comparison Matrix


| Feature             | single_thread       | race                    | not_rate_limit       | rate_limit         |
| ------------------- | ------------------- | ----------------------- | -------------------- | ------------------ |
| Threading           | ❌ No               | ✅ Yes                  | ✅ Yes               | ✅ Yes             |
| Concurrent Requests | ❌ Sequential       | ✅ Parallel             | ✅ Parallel          | ✅ Parallel        |
| Request Counter     | ✅ Yes (safe)       | ✅ Yes                  | ✅ Yes               | ✅ Yes             |
| Counter Thread-Safe | N/A (single thread) | ❌ No (intentional bug) | ✅ Yes (with lock)   | ✅ Yes (with lock) |
| Rate Limiting       | ❌ No               | ❌ No                   | ❌ No                | ✅ Yes (per-IP)    |
| Performance         | Low (1x baseline)   | High (~10x)             | High (~10x)          | High (~10x)        |
| Use Case            | Comparison baseline | Demonstrate problem     | Demonstrate solution | Production ready   |

### 1.2 Code Differences Explained

#### server_single_thread.py - The Baseline

**Purpose:** Establish performance baseline and demonstrate sequential processing

**Key Characteristic:**

```python
# No threading - sequential processing
while True:
    client_socket, address = server_socket.accept()
    handle_request(client_socket, base_dir, address)  # Blocks until complete
    client_socket.close()
```

**Why it exists:** Provides a reference point to measure the performance improvement gained by introducing multithreading.

---

#### server_race.py - The Problem

**Purpose:** Demonstrate what goes wrong without proper synchronization

**Key Characteristic:**

```python
# Intentional race condition - NO LOCK
old_value = request_counter[request_path]
time.sleep(0.001)  # Force thread interleaving
request_counter[request_path] = old_value + 1
```

**Why it exists:** Shows the practical consequences of race conditions. Students can see firsthand how data corruption occurs when multiple threads access shared resources without synchronization.

---

#### server_not_rate_limit.py - The Solution

**Purpose:** Demonstrate proper synchronization with locks

**Key Characteristic:**

```python
# Thread-safe counter with lock
counter_lock = threading.Lock()

with counter_lock:
    request_counter[request_path] += 1  # Atomic operation
```

**Why it exists:** Demonstrates the correct way to handle shared resources in multithreaded environments. Shows how locks prevent race conditions and ensure data integrity.

---

#### server_rate_limit.py - Production Ready

**Purpose:** Add advanced feature (rate limiting) while maintaining thread safety

**Key Characteristics:**

```python
# Per-IP rate limiting with separate locks
rate_limit_data = defaultdict(lambda: {
    'requests': [],
    'lock': threading.Lock()
})

# Thread-safe rate limit check
with client_data['lock']:
    # Check and update rate limit data
```

**Why it exists:** Demonstrates a complete, production-ready implementation with:

- Thread-safe request counting
- Per-IP rate limiting using sliding window algorithm
- Independent locks for better concurrency (per-IP rather than global)

### 1.3 Progressive Learning Path

The four implementations follow a logical learning progression:

1. **single_thread** → Understand the problem (slow performance with sequential processing)
2. **race** → See what happens when multithreading is done wrong
3. **not_rate_limit** → Learn the correct solution (proper synchronization)
4. **rate_limit** → Apply concepts to build advanced features

This progression allows students to:

- Measure the performance benefits of multithreading
- Experience race conditions firsthand
- Understand the necessity of synchronization mechanisms
- Apply thread-safety principles to real-world features

---

## 2. Project Structure

### Command:

```bash
cd ~/Desktop/PR/lab2
ls -la
```

### Files:

* **server_single_thread.py** - Single-threaded baseline server
* **server_not_rate_limit.py** - Multithreaded server with thread-safe counter
* **server_race.py** - Multithreaded server with intentional race condition
* **server_rate_limit.py** - Multithreaded server with rate limiting
* **test_conc.py** - Concurrent requests test script
* **test_spam.py** - Rate limiting test script
* **client.py** - HTTP client from Lab 1
* **content/** - Directory with test files
* **screenshots/** - Documentation screenshots

---

## 3. Server Implementations

### 3.1 Single-Threaded Server (Baseline)

**File:** `server_single_thread.py`

**Key Characteristics:**

* No threading
* Sequential request processing
* Simple counter (no race conditions possible)
* Each request blocks the next one

### 3.2 Multithreaded Server without Rate Limiting

**File:** `server_not_rate_limit.py`

**Key Features:**

* Creates a new thread for each incoming connection
* Thread-safe counter using `threading.Lock()`
* Concurrent request processing
* Directory listing shows request counts per file

**Synchronization:**

```python
counter_lock = threading.Lock()

with counter_lock:
    request_counter[request_path] += 1
```

### 3.3 Server with Race Condition

**File:** `server_race.py`

**Race Condition Implementation:**

```python
old_value = request_counter[request_path]
time.sleep(0.001)
request_counter[request_path] = old_value + 1
```

### 3.4 Server with Rate Limiting

**File:** `server_rate_limit.py`

**Key Features:**

* Thread-safe rate limiting per client IP
* Sliding window algorithm (5 requests/second)
* Returns HTTP 429 when limit exceeded

---

## 4. Performance Comparison: Single-Threaded vs Multithreaded

### 4.1 Single-Threaded Server Test

#### Results:

```
Total time for 10 requests: 10.011s
Average time per request: 5.504s
Throughput: 1.00 requests/second
```

**Analysis:**

* **Total Time:** 10.011 seconds
* **Throughput:** 1.00 req/s
* **Behavior:** Sequential processing

### 4.2 Multithreaded Server Test

#### Results:

```
Total time for 10 requests: 1.008s
Average time per request: 1.006s
Throughput: 9.92 requests/second
```

**Analysis:**

* **Total Time:** 1.008 seconds
* **Throughput:** 9.92 req/s
* **Performance:** **~10x faster** than single-threaded

**Comparison Table:**


| Metric     | Single-Threaded | Multithreaded | Improvement     |
| ---------- | --------------- | ------------- | --------------- |
| Total Time | 10.011s         | 1.008s        | **9.9x faster** |
| Throughput | 1.00 req/s      | 9.92 req/s    | **9.9x higher** |

---

## 5. Race Condition Demonstration

### 5.1 Server with Race Condition

#### Test:

```bash
for i in {1..200}; do curl -s http://localhost:8080/index.html > /dev/null & done
```

#### Result:

```
index.html - 66 requests
```

![Race Condition](screenshots/05_race_condition_browser.png)

**Analysis:**

* **Expected Counter:** 200 requests
* **Actual Counter:** 66 requests
* **Lost Updates:** 134 requests (67%)

**Explanation:**

Multiple threads read and write the counter without synchronization:

1. Thread A reads: `old_value = 10`
2. Thread B reads: `old_value = 10`
3. Thread A writes: `counter = 11`
4. Thread B writes: `counter = 11` (overwrites!)

Result: 2 requests processed, counter only shows 1 increment.

**Server Logs:**

```
Counter for /index.html: 1
Counter for /index.html: 1  ← Same value!
Counter for /index.html: 2
Counter for /index.html: 2  ← Race condition
Counter for /index.html: 1  ← Goes backwards!
```

### 5.2 Server with Lock (Fixed)

#### Test:

```bash
for i in {1..100}; do curl -s http://localhost:8080/index.html > /dev/null & done
```

#### Result:

```
index.html - 100 requests
```

![Counter Fixed](screenshots/06_race_fixed_browser.png)

**Analysis:**

* **Expected Counter:** 100 requests
* **Actual Counter:** 100 requests ✅
* **Lost Updates:** 0 (0%)

**Solution:**

```python
with counter_lock:
    request_counter[request_path] += 1
```

The lock ensures atomic operations - only one thread can modify the counter at a time.

---

## 6. Rate Limiting Implementation

### 6.1 Setup

```bash
sudo ip addr add 127.0.0.2/8 dev lo
```

### 6.2 Rate Limiting Test

#### Test:

```bash
python test_spam.py 127.0.0.1 CLIENT1 & python test_spam.py 127.0.0.2 CLIENT2 &
```

#### Results:

**Client 1 (127.0.0.1):**

```
[CLIENT1] RESULTS:
  Total time: 1.55s
  Successful (200): 4
  Rate limited (429): 26
  Actual rate: 19.37 req/s
```

**Client 2 (127.0.0.2):**

```
[CLIENT2] RESULTS:
  Total time: 1.55s
  Successful (200): 6
  Rate limited (429): 24
  Actual rate: 19.38 req/s
```

**Analysis:**


| Metric             | CLIENT1 (127.0.0.1) | CLIENT2 (127.0.0.2) |
| ------------------ | ------------------- | ------------------- |
| Total Requests     | 30                  | 30                  |
| Successful (200)   | 4                   | 6                   |
| Rate Limited (429) | 26                  | 24                  |
| Success Rate       | 13.3%               | 20.0%               |

**Key Observations:**

* Each IP has **independent rate limiting**
* Both clients sent ~19 req/s (exceeding 5 req/s limit)
* ~80-87% of requests were blocked
* Rate limiting is **thread-safe** for concurrent clients

---

## 7. Technical Implementation Details

### 7.1 Threading Architecture

```python
client_thread = threading.Thread(
    target=handle_request,
    args=(client_socket, base_dir, address),
    daemon=True
)
client_thread.start()
```

**Features:**

* One thread per connection
* Daemon threads (automatic cleanup)
* Non-blocking connection acceptance
* Listen backlog: 100 connections

### 7.2 Synchronization Mechanisms

**Counter Lock:**

```python
counter_lock = threading.Lock()

with counter_lock:
    request_counter[request_path] += 1
```

**Rate Limit Lock (per-IP):**

```python
rate_limit_data = defaultdict(lambda: {
    'requests': [],
    'lock': threading.Lock()
})

with client_data['lock']:
    # Thread-safe rate limit check
```

### 7.3 Rate Limiting Algorithm (Sliding Window)

1. Store timestamp of each request per IP
2. Remove timestamps older than 1 second
3. Count remaining timestamps
4. If count ≥ 5, reject with 429
5. Otherwise, add new timestamp and allow

**Advantages:**

* Smooth rate limiting
* Per-IP isolation
* Thread-safe
* Memory efficient

---

## 8. Performance Metrics Summary

### Concurrency Performance


| Server Type     | 10 Requests Time | Throughput | Speedup  |
| --------------- | ---------------- | ---------- | -------- |
| Single-threaded | 10.011s          | 1.00 req/s | 1x       |
| Multithreaded   | 1.008s           | 9.92 req/s | **9.9x** |

### Counter Accuracy


| Server Type    | Expected | Actual | Accuracy |
| -------------- | -------- | ------ | -------- |
| Race Condition | 200      | 66     | 33%      |
| With Lock      | 100      | 100    | **100%** |

### Rate Limiting Effectiveness


| Client  | Rate Limit | Actual Rate | Blocked |
| ------- | ---------- | ----------- | ------- |
| CLIENT1 | 5 req/s    | 19.37 req/s | 86.7%   |
| CLIENT2 | 5 req/s    | 19.38 req/s | 80.0%   |

## Conclusion

This lab successfully demonstrates multithreading implementation and concurrency control in HTTP servers through four progressive server implementations.

The performance testing shows that multithreading provides significant benefits for I/O-bound operations. The multithreaded server achieved a 9.9x performance improvement over the single-threaded version, processing 10 concurrent requests in 1.008 seconds compared to 10.011 seconds for sequential processing. This demonstrates that concurrent request handling dramatically improves server responsiveness and throughput.

The race condition demonstration clearly illustrates the dangers of unsynchronized access to shared resources. Without proper locking mechanisms, the server lost 67% of request counts due to multiple threads overwriting each other's updates. By implementing threading.Lock() for synchronization, the server achieved 100% counting accuracy, proving that proper synchronization is essential for data integrity in multithreaded applications.
