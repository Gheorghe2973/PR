import socket
import os
import sys
from pathlib import Path
import threading
import time
from collections import defaultdict

MIME_TYPES = {
    '.html': 'text/html',
    '.png': 'image/png',
    '.pdf': 'application/pdf'
}

request_counter = defaultdict(int)
counter_lock = threading.Lock()

def parse_request(request_data):
    lines = request_data.split('\r\n')
    request_line = lines[0].split()
    
    if len(request_line) < 3:
        return None, None, None
    
    method = request_line[0]
    path = request_line[1]
    
    headers = {}
    for line in lines[1:]:
        if ':' in line:
            key, value = line.split(':', 1)
            headers[key.strip()] = value.strip()
    
    return method, path, headers

def generate_directory_listing(dir_path, url_path):
    items = []
    
    if not url_path.startswith('/'):
        url_path = '/' + url_path
    
    try:
        for item in sorted(os.listdir(dir_path)):
            item_path = os.path.join(dir_path, item)
            
            if os.path.isdir(item_path):
                if url_path.endswith('/'):
                    items.append(f'<li><a href="{url_path}{item}/">{item}/</a></li>')
                else:
                    items.append(f'<li><a href="{url_path}/{item}/">{item}/</a></li>')
            else:
                file_key = os.path.join(url_path, item) if not url_path.endswith('/') else url_path + item
                file_key = file_key.replace('//', '/')
                
                with counter_lock:
                    count = request_counter.get(file_key, 0)
                
                if url_path.endswith('/'):
                    items.append(f'<li><a href="{url_path}{item}">{item}</a> - {count} requests</li>')
                else:
                    items.append(f'<li><a href="{url_path}/{item}">{item}</a> - {count} requests</li>')
    except Exception as e:
        return None
    
    html = f"""<!DOCTYPE html>
<html>
<head>
    <title>Directory listing for {url_path}</title>
</head>
<body>
    <h1>Directory listing for {url_path}</h1>
    <hr>
    <ul>
        <li><a href="../">Parent Directory</a></li>
        {''.join(items)}
    </ul>
    <hr>
</body>
</html>"""
    
    return html.encode('utf-8')

def handle_request(client_socket, base_dir, client_address):
    try:
        raw_data = client_socket.recv(4096)
        if not raw_data:
            return
        request_data = raw_data.decode('utf-8', errors='ignore')
    except Exception as e:
        return
    
    try:
        method, path, headers = parse_request(request_data)
        
        if method != 'GET':
            response = b"HTTP/1.1 405 Method Not Allowed\r\n\r\n"
            client_socket.sendall(response)
            return
        
        time.sleep(1)
        
        if path.startswith('/'):
            path = path[1:]
        
        file_path = os.path.join(base_dir, path)
        file_path = os.path.normpath(file_path)
        
        abs_base = os.path.abspath(base_dir)
        abs_file = os.path.abspath(file_path)
        
        if not abs_file.startswith(abs_base):
            response = b"HTTP/1.1 403 Forbidden\r\n\r\nAccess denied"
            client_socket.sendall(response)
            return
        
        if os.path.isdir(file_path):
            html_content = generate_directory_listing(file_path, '/' + path)
            if html_content:
                response_header = (
                    f"HTTP/1.1 200 OK\r\n"
                    f"Content-Type: text/html\r\n"
                    f"Content-Length: {len(html_content)}\r\n"
                    f"\r\n"
                ).encode('utf-8')
                response = response_header + html_content
                client_socket.sendall(response)
            else:
                response = b"HTTP/1.1 404 Not Found\r\n\r\nDirectory not found"
                client_socket.sendall(response)
            return
        
        if not os.path.isfile(file_path):
            response = b"HTTP/1.1 404 Not Found\r\n\r\nFile not found"
            client_socket.sendall(response)
            return
        
        ext = os.path.splitext(file_path)[1].lower()
        
        if ext not in MIME_TYPES:
            response = b"HTTP/1.1 415 Unsupported Media Type\r\n\r\n"
            client_socket.sendall(response)
            return
        
        request_path = '/' + path if path else '/'
        with counter_lock:
            request_counter[request_path] += 1
        
        with open(file_path, 'rb') as f:
            file_content = f.read()
        
        mime_type = MIME_TYPES[ext]
        response_header = (
            f"HTTP/1.1 200 OK\r\n"
            f"Content-Type: {mime_type}\r\n"
            f"Content-Length: {len(file_content)}\r\n"
            f"\r\n"
        ).encode('utf-8')
        
        client_socket.sendall(response_header + file_content)
        
    except Exception as e:
        try:
            response = b"HTTP/1.1 500 Internal Server Error\r\n\r\n"
            client_socket.sendall(response)
        except:
            pass
    finally:
        try:
            client_socket.close()
        except:
            pass

def start_server(port, base_dir):
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    
    server_socket.bind(('0.0.0.0', port))
    server_socket.listen(50)
    
    print(f"Server started on port {port}")
    print(f"Serving directory: {os.path.abspath(base_dir)}")
    print("Press Ctrl+C to stop\n")
    
    try:
        while True:
            client_socket, address = server_socket.accept()
            
            client_thread = threading.Thread(
                target=handle_request,
                args=(client_socket, base_dir, address),
                daemon=True
            )
            client_thread.start()
            
    except KeyboardInterrupt:
        print("\nShutting down server...")
    finally:
        server_socket.close()

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python server_not_rate_limit.py <directory>")
        sys.exit(1)
    
    directory = sys.argv[1]
    
    if not os.path.isdir(directory):
        print(f"Error: {directory} is not a valid directory")
        sys.exit(1)
    
    PORT = 8080
    start_server(PORT, directory)