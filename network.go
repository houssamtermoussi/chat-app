package main

import (
	"log"
	"net"
)

const defaultPort = "8080"

// localIPv4Addresses retourne les IP locales (Wi‑Fi, Ethernet) utilisables par d'autres appareils.
func localIPv4Addresses() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}

func printListenURLs(port string) {
	log.Println("Serveur accessible sur ce PC :")
	log.Printf("  → http://localhost:%s", port)

	ips := localIPv4Addresses()
	if len(ips) == 0 {
		log.Println("Aucune IP locale détectée — vérifie ta connexion Wi‑Fi / Ethernet.")
		return
	}

	log.Println("Pour un autre appareil (même Wi‑Fi), ouvre une de ces adresses :")
	seen := make(map[string]bool)
	for _, ip := range ips {
		if seen[ip] {
			continue
		}
		seen[ip] = true
		log.Printf("  → http://%s:%s", ip, port)
	}
	log.Println("Les deux appareils doivent utiliser la même adresse (pas localhost sur le téléphone).")
}
