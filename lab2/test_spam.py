import socket
import time

def spam_test(server_host, server_port, client_name, source_ip):
    print(f"[{client_name}] Sending 30 requests from {source_ip} to {server_host}:{server_port}...")
    
    successful = 0
    rate_limited = 0
    start = time.time()
    
    for i in range(30):
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            s.settimeout(2)
            s.bind((source_ip, 0))  
            s.connect((server_host, server_port))
            
            request = "GET /index.html HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
            s.sendall(request.encode())
            
            response = s.recv(1024).decode('utf-8', errors='ignore')
            
            if '200 OK' in response:
                successful += 1
                print(f"  [{client_name}] Request {i+1}: 200 OK")
            elif '429' in response:
                rate_limited += 1
                print(f"  [{client_name}] Request {i+1}: 429 RATE LIMITED")
            
            s.close()
        except Exception as e:
            print(f"  [{client_name}] Request {i+1}: Error - {e}")
        
        time.sleep(0.05)
    
    duration = time.time() - start
    
    print(f"\n[{client_name}] RESULTS:")
    print(f"  Total time: {duration:.2f}s")
    print(f"  Successful (200): {successful}")
    print(f"  Rate limited (429): {rate_limited}")
    print(f"  Actual rate: {30/duration:.2f} req/s")
    print()

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) > 1:
        server_host = sys.argv[1]
        source_ip = sys.argv[2] if len(sys.argv) > 2 else "127.0.0.1"
        client_name = sys.argv[3] if len(sys.argv) > 3 else "CLIENT"
    else:
        server_host = "127.0.0.1"
        source_ip = "127.0.0.1"
        client_name = "CLIENT1"
    
    print("=" * 80)
    print(f"SPAM TEST from {source_ip} to {server_host}")
    print("=" * 80)
    
    spam_test(server_host, 8080, client_name, source_ip)