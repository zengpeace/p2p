detail:(client - NATA - transit - NATB - server)
1. server to transit long connection
send:
cmd = 101
cpuId = "own cpuId"
return:
cmd = 101
result = int8(0 success, -1 fail)

2. client to transit short connection
send:
cmd = 102
cpuId = "the cpuId of server which client want to connect"
return:
cmd = 102
result = int8(0 success, -1 fail)
NATB value

3. transit to server
send:
cmd = 103
NATA value
return:
no

4. server send data to NATA
send:
cmd = 104
return:
no

5. client connect to NATB, send data as usual
send:
cmd = 201
real communicate data
