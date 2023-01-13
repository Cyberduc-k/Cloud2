import http.server
import socketserver

class StaticFileHttpRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.path = f"/static{self.path}"
        print(self.path)
        return http.server.SimpleHTTPRequestHandler.do_GET(self)

PORT = 8081
handler = StaticFileHttpRequestHandler

with socketserver.TCPServer(("", PORT), handler) as httpd:
    print("Server started at localhost:" + str(PORT))
    httpd.serve_forever()
