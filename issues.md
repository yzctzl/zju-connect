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


-------------------------------------
2026/01/10 18:30:54 KeepAlive: OK                                                                                             
2026/01/10 18:31:54 KeepAlive: OK                              
2026/01/10 18:32:54 KeepAlive: OK                           
2026/01/10 18:33:54 KeepAlive: OK                                                                                             
2026/01/10 18:34:54 KeepAlive: OK                                                                                             
2026/01/10 18:35:54 KeepAlive: OK                                                                                             
2026/01/10 18:36:54 KeepAlive: OK                                                                                             
2026/01/10 18:37:54 KeepAlive: OK                                                                                             
2026/01/10 18:38:54 KeepAlive: OK                                                                                             
2026/01/10 18:39:54 KeepAlive: OK                                                                                             2026/01/10 18:40:54 KeepAlive: OK                              
2026/01/10 18:41:54 KeepAlive: OK                                                                                             2026/01/10 18:42:54 KeepAlive: OK
2026/01/10 18:43:52 Error occurred while receiving (attempt 1): read tcp 172.19.14.48:43966->221.176.159.68:443: read: connect
ion reset by peer                                                                                                             2026/01/10 18:43:52 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:43:52 Waiting 2s before reconnect...                                                                            
2026/01/10 18:43:54 Error occurred while sending (attempt 1): write tcp 172.19.14.48:43962->221.176.159.68:443: write: connection reset by peer                                                                                                             
2026/01/10 18:43:54 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:43:54 Waiting 2s before reconnect...             
2026/01/10 18:43:54 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    2026/01/10 18:43:54 Waiting 4s before reconnect...   
2026/01/10 18:43:56 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    2026/01/10 18:43:56 Waiting 4s before reconnect...                                                                            
2026/01/10 18:43:58 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    2026/01/10 18:43:58 Waiting 8s before reconnect...                                                                            
2026/01/10 18:44:00 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    2026/01/10 18:44:00 Waiting 8s before reconnect...                                                                            
2026/01/10 18:44:04 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:42355->202.196.96.131:53: i/o tim
eout (consecutive failures: 1)                                                                                                
2026/01/10 18:44:06 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:44:06 Waiting 16s before reconnect...                                                                           
2026/01/10 18:44:08 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:44:08 Waiting 16s before reconnect...                                                                           
2026/01/10 18:44:22 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...
2026/01/10 18:44:08 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying... 18:44:54 [280/1905]
2026/01/10 18:44:08 Waiting 16s before reconnect...                                                                           
2026/01/10 18:44:22 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    2026/01/10 18:44:22 Waiting 32s before reconnect...                                                                           
2026/01/10 18:44:24 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    2026/01/10 18:44:24 Waiting 32s before reconnect...                                                                           
2026/01/10 18:44:54 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 18:44:54 TLS: connected to: 221.176.159.68:443                                                                     
2026/01/10 18:44:55 RecvConn reconnected successfully                                                                         
2026/01/10 18:44:56 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 18:44:56 TLS: connected to: 221.176.159.68:443                                                                     
2026/01/10 18:44:57 SendConn reconnected successfully                                                                         
2026/01/10 18:45:04 KeepAlive: recovered after 1 failures                                                                     
2026/01/10 18:45:04 KeepAlive: OK                                                                                             
2026/01/10 18:46:04 KeepAlive: OK                                                                                             
2026/01/10 18:46:04 KeepWebSessionAlive: OK                                                                                   2026/01/10 18:46:36 Error occurred while receiving (attempt 1): read tcp 172.19.14.48:34680->221.176.159.68:443: read: connect
ion reset by peer                                                                                                             2026/01/10 18:46:36 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:46:36 Waiting 2s before reconnect...                                                                            
2026/01/10 18:46:38 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    2026/01/10 18:46:38 Waiting 4s before reconnect...                                                                            
2026/01/10 18:46:42 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:46:42 Waiting 8s before reconnect...                                                                            2026/01/10 18:46:50 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:46:50 Waiting 16s before reconnect...                                                                           
2026/01/10 18:46:51 Error occurred while sending (attempt 1): write tcp 172.19.14.48:34682->221.176.159.68:443: write: connect
ion reset by peer                                                                                                             2026/01/10 18:46:51 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:46:51 Waiting 2s before reconnect...                                                                            2026/01/10 18:46:53 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:46:53 Waiting 4s before reconnect...                                                                            2026/01/10 18:46:57 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:46:57 Waiting 8s before reconnect...                                                                            2026/01/10 18:47:05 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:47:05 Waiting 16s before reconnect...                                                                           
2026/01/10 18:47:06 RecvConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:47:06 Waiting 32s before reconnect...                                                                           
2026/01/10 18:47:14 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:49798->202.196.96.131:53: i/o tim
eout (consecutive failures: 1)                                                                                                
2026/01/10 18:47:21 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying...                    
2026/01/10 18:47:21 Waiting 32s before reconnect...                                        
2026/01/10 18:47:21 SendConn failed: dial tcp 221.176.159.68:443: connect: connection refused. Retrying... 18:47:53 [240/1905]
2026/01/10 18:47:21 Waiting 32s before reconnect...                                                                           
2026/01/10 18:47:38 Socket: connected to: 221.176.159.68:443                                                                  2026/01/10 18:47:38 TLS: connected to: 221.176.159.68:443                                                                     
2026/01/10 18:47:38 RecvConn failed: EOF. Retrying...                                                                         2026/01/10 18:47:38 Waiting 1m4s before reconnect...                                                                          
2026/01/10 18:47:53 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 18:47:53 TLS: connected to: 221.176.159.68:443                                                                     
2026/01/10 18:47:53 SendConn failed: EOF. Retrying...                                                                         
2026/01/10 18:47:53 Waiting 1m4s before reconnect...                                                                          
2026/01/10 18:48:24 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:46511->202.196.96.131:53: i/o tim
eout (consecutive failures: 2)                                                                                                
2026/01/10 18:48:42 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 18:48:42 TLS: connected to: 221.176.159.68:443                                                                     
2026/01/10 18:48:42 RecvConn failed: EOF. Retrying...                                                                         
2026/01/10 18:48:42 Waiting 2m8s before reconnect...                                                                          2026/01/10 18:48:57 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 18:48:57 TLS: connected to: 221.176.159.68:443                                                                     2026/01/10 18:48:57 SendConn failed: EOF. Retrying...                                                                         
2026/01/10 18:48:57 Waiting 2m8s before reconnect...                                                                          
[SOCKS5] 2026/01/10 18:49:09 [E]: server: writeto tcp4 10.11.137.79:45085->10.91.28.4:21834: readfrom tcp 127.0.0.1:10807->127.0.0.1:40130: splice: connection timed out                                                                                    
2026/01/10 18:49:34 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:60986->202.196.96.131:53: i/o tim
eout (consecutive failures: 3)                                                                                                2026/01/10 18:50:44 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:40042->202.196.96.131:53: i/o tim
eout (consecutive failures: 4)                                                                                                
2026/01/10 18:50:50 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 18:50:50 TLS: connected to: 221.176.159.68:443                                                                     2026/01/10 18:50:50 RecvConn failed: EOF. Retrying...                                                                         
2026/01/10 18:50:50 Waiting 4m16s before reconnect...                                                                         2026/01/10 18:51:05 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 18:51:05 TLS: connected to: 221.176.159.68:443                                                                     2026/01/10 18:51:05 SendConn failed: EOF. Retrying...                                                                         
2026/01/10 18:51:05 Waiting 4m16s before reconnect...                                                                         2026/01/10 18:51:54 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:35757->202.196.96.131:53: i/o tim
eout (consecutive failures: 5)                                                                                                
2026/01/10 18:51:54 KeepAlive: too many consecutive failures, attempting to recreate resolver...                              
2026/01/10 18:51:54 KeepAlive: resolver recreated successfully                                                                
2026/01/10 18:53:04 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:54943->202.196.96.131:53: i/o tim
eout (consecutive failures: 1)                                                                                                
2026/01/10 18:54:14 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:44529->202.196.96.131:53: i/o tim
eout (consecutive failures: 2)  
[SOCKS5] 2026/01/10 19:14:17 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.919:15:07 [79/1905]
ect: connection timed out                                      
2026/01/10 19:14:23 10.91.28.4:21834 -> VPN                                                                                   [SOCKS5] 2026/01/10 19:15:07 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                                                                                     2026/01/10 19:15:07 Socket: connected to: 221.176.159.68:443   
2026/01/10 19:15:07 TLS: connected to: 221.176.159.68:443
2026/01/10 19:15:07 RecvConn failed: EOF. Retrying...       
2026/01/10 19:15:07 Waiting 5m0s before reconnect...     
2026/01/10 19:15:09 10.91.28.4:21834 -> VPN                 
2026/01/10 19:15:14 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:56396->202.196.96.131:53: i/o tim
eout (consecutive failures: 5)                                 
2026/01/10 19:15:14 KeepAlive: too many consecutive failures, attempting to recreate resolver...
2026/01/10 19:15:14 KeepAlive: resolver recreated successfully
2026/01/10 19:15:22 Socket: connected to: 221.176.159.68:443                                                                  
2026/01/10 19:15:22 TLS: connected to: 221.176.159.68:443                                                                     2026/01/10 19:15:22 SendConn failed: EOF. Retrying...          
2026/01/10 19:15:22 Waiting 5m0s before reconnect...                                                                          [SOCKS5] 2026/01/10 19:15:52 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                                                                                     
2026/01/10 19:15:55 10.91.28.4:21834 -> VPN                                                                                   2026/01/10 19:16:24 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:37467->202.196.96.131:53: i/o tim
eout (consecutive failures: 1)                                 
2026/01/10 19:16:25 KeepWebSessionAlive: OK                                                                                   [SOCKS5] 2026/01/10 19:16:37 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                                                                                     
2026/01/10 19:16:41 10.91.28.4:21834 -> VPN                   
[SOCKS5] 2026/01/10 19:17:22 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out     
2026/01/10 19:17:27 10.91.28.4:21834 -> VPN                                                                                   2026/01/10 19:17:35 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:60674->202.196.96.131:53: i/o tim
eout (consecutive failures: 2)                                                                                                [SOCKS5] 2026/01/10 19:18:11 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                                                                                     2026/01/10 19:18:13 10.91.28.4:21834 -> VPN                                                                                   
2026/01/10 19:18:45 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:50386->202.196.96.131:53: i/o tim
eout (consecutive failures: 3)                                                                                                
[SOCKS5] 2026/01/10 19:18:56 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                                                                                     
2026/01/10 19:18:59 10.91.28.4:21834 -> VPN
2026/01/10 19:18:59 10.91.28.4:21834 -> VPN                                                                 19:20:07 [40/1905][SOCKS5] 2026/01/10 19:19:41 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                      
2026/01/10 19:19:45 10.91.28.4:21834 -> VPN                                                                                   2026/01/10 19:19:55 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:37755->202.196.96.131:53: i/o tim
eout (consecutive failures: 4)                                                                                                2026/01/10 19:20:07 Socket: connected to: 221.176.159.68:443   
2026/01/10 19:20:07 TLS: connected to: 221.176.159.68:443
2026/01/10 19:20:07 RecvConn failed: EOF. Retrying...       
2026/01/10 19:20:07 Waiting 5m0s before reconnect...     
2026/01/10 19:20:22 Socket: connected to: 221.176.159.68:443
2026/01/10 19:20:22 TLS: connected to: 221.176.159.68:443
2026/01/10 19:20:22 SendConn failed: EOF. Retrying...       
2026/01/10 19:20:22 Waiting 5m0s before reconnect...     
2026/01/10 19:20:25 10.91.28.4:21834 -> VPN          
[SOCKS5] 2026/01/10 19:20:26 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                                                                                     2026/01/10 19:20:31 10.91.28.4:21834 -> VPN                    
2026/01/10 19:21:05 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:42431->202.196.96.131:53: i/o timeout (consecutive failures: 5)
2026/01/10 19:21:05 KeepAlive: too many consecutive failures, attempting to recreate resolver...                              
2026/01/10 19:21:05 KeepAlive: resolver recreated successfully                                                                [SOCKS5] 2026/01/10 19:21:11 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                      
2026/01/10 19:21:17 10.91.28.4:21834 -> VPN                                                                                   [SOCKS5] 2026/01/10 19:22:00 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                                                                                     
2026/01/10 19:22:02 10.91.28.4:21834 -> VPN                   
2026/01/10 19:22:15 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:58647->202.196.96.131:53: i/o timeout (consecutive failures: 1)
[SOCKS5] 2026/01/10 19:22:41 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out
[SOCKS5] 2026/01/10 19:22:45 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out     
2026/01/10 19:22:49 10.91.28.4:21834 -> VPN                                                                                   2026/01/10 19:23:25 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:59282->202.196.96.131:53: i/o tim
eout (consecutive failures: 2)                                 
[SOCKS5] 2026/01/10 19:23:30 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: conn
ect: connection timed out                                      
2026/01/10 19:23:35 10.91.28.4:21834 -> VPN
[SOCKS5] 2026/01/10 19:24:15 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out
2026/01/10 19:24:20 10.91.28.4:21834 -> VPN
2026/01/10 19:24:35 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:60451->202.196.96.131:53: i/o timeout (consecutive failures: 3)
[SOCKS5] 2026/01/10 19:25:05 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out
2026/01/10 19:25:07 10.91.28.4:21834 -> VPN
2026/01/10 19:25:07 Socket: connected to: 221.176.159.68:443
2026/01/10 19:25:07 TLS: connected to: 221.176.159.68:443
2026/01/10 19:25:07 RecvConn failed: EOF. Retrying...
2026/01/10 19:25:07 Waiting 5m0s before reconnect...
2026/01/10 19:25:22 Socket: connected to: 221.176.159.68:443
2026/01/10 19:25:22 TLS: connected to: 221.176.159.68:443
2026/01/10 19:25:22 SendConn failed: EOF. Retrying...
2026/01/10 19:25:22 Waiting 5m0s before reconnect...
2026/01/10 19:25:45 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:57963->202.196.96.131:53: i/o timeout (consecutive failures: 4)
[SOCKS5] 2026/01/10 19:25:50 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out
2026/01/10 19:25:53 10.91.28.4:21834 -> VPN
[SOCKS5] 2026/01/10 19:26:35 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out
2026/01/10 19:26:39 10.91.28.4:21834 -> VPN
2026/01/10 19:26:55 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:55569->202.196.96.131:53: i/o timeout (consecutive failures: 5)
2026/01/10 19:26:55 KeepAlive: too many consecutive failures, attempting to recreate resolver...
2026/01/10 19:26:55 KeepAlive: resolver recreated successfully
[SOCKS5] 2026/01/10 19:27:20 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out
^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[A^[[B^[[B^[[B^[[B^[[B^[[B2026/01/10 19:27:25 10.91.28.4:21834 -> VPN
2026/01/10 19:28:05 KeepAlive: lookup www.baidu.com on 127.0.0.53:53: read udp4 10.11.137.79:41783->202.196.96.131:53: i/o timeout (consecutive failures: 1)
[SOCKS5] 2026/01/10 19:28:09 [E]: server: connect to 10.91.28.4:21834 failed, dial tcp4 10.11.137.79:0->10.91.28.4:21834: connect: connection timed out
2026/01/10 19:28:10 10.91.28.4:21834 -> VPN
^C2026/01/10 19:28:33 Shutdown ZJU-Connect ......
2026/01/10 19:28:33 Exec func on terminal: Close Tun Device
2026/01/10 19:28:33 Exec func on terminal  Close Tun Device success
2026/01/10 19:28:33 Shutdown ZJU-Connect success, Bye~
