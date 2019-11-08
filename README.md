# rdsb-ps
A Public Server Stats proxy for the Roller Derby Scoreboard (at https://github.com/rollerderby/scoreboard )

The Roller Derby Scoreboard (previously known as the Carolina Scoreboard) provides a lot of additional statistical data, and allows electronic operation of lineup, penalty tracking etc.
However: there is no security on the Scoreboard [the webserver trusts anyone who can talk to it]. 

This means that it has been difficult, to date, to safely provide access to live stats to spectators.

This simple proxy tries to solve this by reexporting the scoreboard stats on its own, read-only webserver.
(We intend that the proxy be deployed on a host with access to the scoreboard network (secured) *and* an additional public network: the webserver
it runs binds to all interfaces by default, so it should be visible to everyone.)

The present version of the scoreboard is commandline only. It takes 1 argument, the URL (including port) of the scoreboard to register with (it 
defaults to localhost:8000 in the assumption that you're running it on the scoreboard machine.)

invoke as:

rdsp-ps -addr=address:port 

or 

rdsp-ps 

(for the default localhost:8000)

The webserver runs on port 80, and redirects to index.html if you connect to it with no path.

At present, we provide two public statistics pages:
  whiteboard.html : a simple display of penalties per team
  img/scores.png : a simple graph of team scores over time (by jam)
  
  

Note: if you're building this project from source, we recommend that you use:
  go build -ldflags="-s -w" 
and then compress the binary with UPX
  upx rdsb-ps 
as this will produce an output binary at almost 25% of the default compiler output.
  
  
  

