2026/01/09 15:53:48 KeepAlive: OK
2026/01/09 15:54:48 KeepAlive: OK
2026/01/09 15:54:55 Error occurred while sending, retrying: write tcp 172.19.14.48:60220->221.176.159.68:443: write: connection reset by peer
panic: dial tcp 221.176.159.68:443: connect: connection refused

goroutine 51 [running]:
github.com/mythologyli/zju-connect/stack/tun.(*Stack).processIPV4TCP(0xc000244a40, {0xc000187180, 0x88, 0x578}, {0xc000187194, 0x74, 0x564})
        github.com/mythologyli/zju-connect/stack/tun/stack.go:154 +0x3dd
github.com/mythologyli/zju-connect/stack/tun.(*Stack).processIPV4(0xc000244a40, {0xc000187180, 0x88, 0x578})
        github.com/mythologyli/zju-connect/stack/tun/stack.go:130 +0x788
github.com/mythologyli/zju-connect/stack/tun.(*Stack).Run(0xc000244a40)
        github.com/mythologyli/zju-connect/stack/tun/stack.go:83 +0x10e
created by main.main in goroutine 1
        github.com/mythologyli/zju-connect/main_tun.go:212 +0x195b

-----------------------------------------------------
2026/01/09 11:30:11 www.henu.edu.cn -> 172.31.7.4                                                                                   
2026/01/09 11:30:11 172.31.7.4:443 -> VPN                         
2026/01/09 11:31:09 KeepAlive: OK                                                                                                   
2026/01/09 11:32:09 KeepAlive: OK                                 
2026/01/09 11:33:09 KeepAlive: OK                                                                                                   
2026/01/09 11:34:06 Error occurred while sending, retrying: write tcp 172.19.14.48:44676->221.176.159.68:443: write: connection rese
t by peer                                                                                                                           
2026/01/09 11:34:06 Socket: connected to: 221.176.159.68:443      
2026/01/09 11:34:06 TLS: connected to: 221.176.159.68:443                                                                           
2026/01/09 11:34:06 Error occurred while receiving, retrying: read tcp 172.19.14.48:44678->221.176.159.68:443: read: connection rese
t by peer                                                                                                                           
panic: dial tcp 221.176.159.68:443: connect: connection refused   
                                                                  
goroutine 37 [running]:                                           
github.com/mythologyli/zju-connect/stack/tun.(*Stack).Run.func1()
        github.com/mythologyli/zju-connect/stack/tun/stack.go:45 +0x178                                                             
created by github.com/mythologyli/zju-connect/stack/tun.(*Stack).Run in goroutine 14
        github.com/mythologyli/zju-connect/stack/tun/stack.go:40 +0x99 

----------------------------------------------------------
2026/01/09 09:58:08 KeepAlive: OK                                 
2026/01/09 09:59:08 KeepAlive: OK                                                                                                   
2026/01/09 10:00:08 KeepAlive: OK                                                                                                   
2026/01/09 10:01:08 KeepAlive: OK                                                                                                   
2026/01/09 10:01:34 Error occurred while receiving, retrying: EOF                                                                   
2026/01/09 10:01:34 Socket: connected to: 221.176.159.68:443                                                                        
2026/01/09 10:01:34 TLS: connected to: 221.176.159.68:443                                                                           
panic: unexpected recv handshake reply                                                                                              
                                                                                                                                    
goroutine 55 [running]:                                                                                                             
github.com/mythologyli/zju-connect/stack/tun.(*Stack).Run.func1()                                                                   
        github.com/mythologyli/zju-connect/stack/tun/stack.go:45 +0x178
created by github.com/mythologyli/zju-connect/stack/tun.(*Stack).Run in goroutine 32
        github.com/mythologyli/zju-connect/stack/tun/stack.go:40 +0x99
