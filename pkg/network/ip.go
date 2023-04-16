package network

import (
	"net"
)

var localIP = ""

func LocalIP() string {
	if localIP == "" {
		netInterfaces, err := net.Interfaces()
		if err != nil {
			//logger.Errorf("get Interfaces failed,err:%+v", err)
			return ""
		}

		for i := 0; i < len(netInterfaces); i++ {
			if ((netInterfaces[i].Flags & net.FlagUp) != 0) && ((netInterfaces[i].Flags & net.FlagLoopback) == 0) {
				addrs, err := netInterfaces[i].Addrs()
				if err != nil {
					//logger.Errorf("get InterfaceAddress failed,err:%+v", err)
					return ""
				}
				for _, address := range addrs {
					if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
						localIP = ipnet.IP.String()
						break
					}
				}
			}
		}

		if len(localIP) > 0 {
			//logger.Infof("Local IP:%s", localIP)
		}
	}
	return localIP
}
