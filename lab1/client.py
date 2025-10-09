import socket
import sys
import os

def parse_response(response_data):
    parts = response_data.split(b'\r\n\r\n', 1)
    
    if len(parts) < 2:
        return None, None, None
    
    header_data = parts[0].decode('utf-8')
    body = parts[1]
    
    lines = header_data.split('\r\n')
    status_line = lines[0]
    status_parts = status_line.split(' ', 2)
    
    if len(status_parts) < 2:
        return None, None, None
    
    status_code = int(status_parts[1])
    
    headers = {}
    for line in lines[1:]:
        if ':' in line:
            key, value = line.split(':', 1)
            headers[key.strip()] = value.strip()
    
    return status_code, headers, body

def send_request(host, port, path):
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    
    try:
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
        
        return response_data
        
    finally:
        client_socket.close()

def get_file_extension(path, content_type):
    if '.' in path:
        ext = os.path.splitext(path)[1]
        if ext:
            return ext
    
    if content_type:
        if 'html' in content_type:
            return '.html'
        elif 'png' in content_type:
            return '.png'
        elif 'pdf' in content_type:
            return '.pdf'
    
    return ''

def main():
    if len(sys.argv) != 5:
        print("Usage: python client.py server_host server_port url_path directory")
        print("Example: python client.py localhost 8080 /index.html downloads")
        sys.exit(1)
    
    host = sys.argv[1]
    port = int(sys.argv[2])
    path = sys.argv[3]
    save_dir = sys.argv[4]
    
    if not path.startswith('/'):
        path = '/' + path
    
    if not os.path.exists(save_dir):
        os.makedirs(save_dir)
    
    print(f"Requesting: http://{host}:{port}{path}")
    
    try:
        response_data = send_request(host, port, path)
    except Exception as e:
        print(f"Error connecting to server: {e}")
        sys.exit(1)
    
    status_code, headers, body = parse_response(response_data)
    
    if status_code is None:
        print("Error: Invalid response from server")
        sys.exit(1)
    
    print(f"Status Code: {status_code}")
    print(f"Content-Type: {headers.get('Content-Type', 'unknown')}")
    print(f"Content-Length: {headers.get('Content-Length', 'unknown')}")
    print()
    
    if status_code != 200:
        print(f"Error: Server returned status code {status_code}")
        print(body.decode('utf-8', errors='ignore'))
        sys.exit(1)
    
    content_type = headers.get('Content-Type', '')
    
    if 'text/html' in content_type:
        print("HTML Content:")
        print("-" * 80)
        print(body.decode('utf-8'))
        print("-" * 80)
        
    elif 'image/png' in content_type or 'application/pdf' in content_type:
        filename = os.path.basename(path)
        if not filename:
            ext = get_file_extension(path, content_type)
            filename = f"downloaded_file{ext}"
        
        save_path = os.path.join(save_dir, filename)
        
        with open(save_path, 'wb') as f:
            f.write(body)
        
        print(f"File saved to: {save_path}")
        print(f"File size: {len(body)} bytes")
    
    else:
        print(f"Unknown content type: {content_type}")
        print("Body preview:")
        print(body[:500])

if __name__ == "__main__":
    main()