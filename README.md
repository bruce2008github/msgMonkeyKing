
*****************************************************************************************************
Note:
   
   The msgMonkeyKing is useful tool to do the TCP network server application testing.

   When msgMonkeyKing receive the client message and it will do some work,then forword to the server.

   It can simulate the TCP network delay and packet fragmentation and assembly.   

   The msgMonkeyKing can handle thousands of client TCP connections concurrently.

   It's written in the go programming language
   (golang: http://en.wikipedia.org/wiki/Go_(programming_language)).


*****************************************************************************************************
Build:

   1) Make sure you have installed golang(golang.org) 

   2) run make command

*****************************************************************************************************
Usage:
   
   1) launch your tcp application server 

   2) launch the msgMonkeyKing with the specified the server address info(ip/port) 
      and msgMonkeyKing's listen port
      (run ./msgMonekyKing -help to see the detail info)

   3) run your client to connect to the msgMonkeyKing as your tcp application server

   4) If you need to run a lot of concurrent connections testing, 
      Please tune your linux server at first, and check the system limit(ulimit -a)


