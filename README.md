Note:
   
   When msgMonkeyKing receive the client message and it will do some work,then forword to the server.

   It can simulate the TCP network delay and packet fragmentation and assembly.

   The msgMonkeyKing is useful to testing the TCP network server application.

   It can handle thousands of client TCP connections concurrently.

Build:

   run make command

Usage:
   
   1) launch your tcp application server 

   2) launch the msgMonkeyKing with the specified the server address info(ip/port) 
      and msgMonkeyKing's listen port
      (run ./msgMonekyKing -help to see the detail info)

   3) run your client to connect to the msgMonkeyKing as your tcp application server


