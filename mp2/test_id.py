import socket, struct

s = socket.socket()
s.connect(('127.0.0.1', 38449))
d = s.recv(4)
print struct.unpack('>i', d)