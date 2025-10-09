import socket
import os
import sys
from pathlib import Path

MIME_TYPES = {
    '.html': 'text/html',
    '.png': 'image/png',
    '.pdf': 'application/pdf'
}

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
                if url_path.endswith('/'):
                    items.append(f'<li><a href="{url_path}{item}">{item}</a></li>')
                else:
                    items.append(f'<li><a href="{url_path}/{item}">{item}</a></li>')
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

def handle_request(client_socket, base_dir):
    try:
        request_data = client_socket.recv(4096).decode('utf-8')
        
        if not request_data:
            return
        
        print(f"Received request:\n{request_data[:200]}...")
        
        method, path, headers = parse_request(request_data)
        
        print(f"Parsed: method={method}, path={path}")
        
        if method != 'GET':
            response = b"HTTP/1.1 405 Method Not Allowed\r\n\r\n"
            client_socket.sendall(response)
            return
        
        if path.startswith('/'):
            path = path[1:]
        
        file_path = os.path.join(base_dir, path)
        file_path = os.path.normpath(file_path)
        
        print(f"Looking for file: {file_path}")
        print(f"Is directory: {os.path.isdir(file_path)}")
        print(f"Is file: {os.path.isfile(file_path)}")
        
        abs_base = os.path.abspath(base_dir)
        abs_file = os.path.abspath(file_path)
        
        if not abs_file.startswith(abs_base):
            response = b"HTTP/1.1 403 Forbidden\r\n\r\nAccess denied"
            client_socket.sendall(response)
            print(f"403 - Forbidden: {abs_file} not in {abs_base}")
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
            print(f"404 - File not found: {file_path}")
            return
        
        ext = os.path.splitext(file_path)[1].lower()
        
        if ext not in MIME_TYPES:
            response = b"HTTP/1.1 415 Unsupported Media Type\r\n\r\n"
            client_socket.sendall(response)
            print(f"415 - Unsupported file type: {ext}")
            return
        
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
        print(f"200 OK - Served: {file_path} ({mime_type})")
        
    except Exception as e:
        print(f"Error handling request: {e}")
        import traceback
        traceback.print_exc()
        try:
            response = b"HTTP/1.1 500 Internal Server Error\r\n\r\n"
            client_socket.sendall(response)
        except:
            pass

def start_server(port, base_dir):
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    
    server_socket.bind(('0.0.0.0', port))
    server_socket.listen(5)
    
    print(f"Server started on port {port}")
    print(f"Serving directory: {os.path.abspath(base_dir)}")
    print("Press Ctrl+C to stop\n")
    
    try:
        while True:
            client_socket, address = server_socket.accept()
            print(f"Connection from {address}")
            
            handle_request(client_socket, base_dir)
            
            client_socket.close()
            
    except KeyboardInterrupt:
        print("\nShutting down server...")
    finally:
        server_socket.close()

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python server.py <directory>")
        sys.exit(1)
    
    directory = sys.argv[1]
    
    if not os.path.isdir(directory):
        print(f"Error: {directory} is not a valid directory")
        sys.exit(1)
    
    PORT = 8080
    start_server(PORT, directory)