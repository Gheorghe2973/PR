import socket
import time
from concurrent.futures import ThreadPoolExecutor, as_completed

def send_request(host, port, path, request_id):
    start_time = time.time()
    
    try:
        client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        client_socket.connect((host, port))
        
        request = (
            f"GET {path} HTTP/1.1\r\n"
            f"Host: {host}\r\n"
            f"Connection: close\r\n"
            f"\r\n"
        ).encode('utf-8')
        
        client_socket.sendall(request)
        
        response_data = b""
        while True:
            chunk = client_socket.recv(4096)
            if not chunk:
                break
            response_data += chunk
        
        client_socket.close()
        
        end_time = time.time()
        elapsed = end_time - start_time
        
        status_line = response_data.split(b'\r\n')[0].decode('utf-8')
        
        return {
            'id': request_id,
            'status': status_line,
            'time': elapsed,
            'start': start_time,
            'end': end_time
        }
        
    except Exception as e:
        return {
            'id': request_id,
            'error': str(e),
            'time': time.time() - start_time
        }

def test_concurrent_requests(host, port, path, num_requests):
    print(f"Testing {num_requests} concurrent requests to http://{host}:{port}{path}")
    print("-" * 80)
    
    overall_start = time.time()
    
    with ThreadPoolExecutor(max_workers=num_requests) as executor:
        futures = [
            executor.submit(send_request, host, port, path, i) 
            for i in range(num_requests)
        ]
        
        results = []
        for future in as_completed(futures):
            result = future.result()
            results.append(result)
            
            if 'error' in result:
                print(f"Request {result['id']}: ERROR - {result['error']}")
            else:
                print(f"Request {result['id']}: {result['status']} - {result['time']:.3f}s")
    
    overall_end = time.time()
    total_time = overall_end - overall_start
    
    print("-" * 80)
    print(f"Total time for {num_requests} requests: {total_time:.3f}s")
    print(f"Average time per request: {sum(r['time'] for r in results) / len(results):.3f}s")
    print(f"Throughput: {num_requests / total_time:.2f} requests/second")
    
    return results, total_time

if __name__ == "__main__":
    HOST = "localhost"
    PORT = 8080
    PATH = "/index.html"
    NUM_REQUESTS = 10
    
    print("=" * 80)
    print("CONCURRENT REQUESTS TEST")
    print("=" * 80)
    print()
    
    results, total_time = test_concurrent_requests(HOST, PORT, PATH, NUM_REQUESTS)
    
    print()
    print("Test completed!")